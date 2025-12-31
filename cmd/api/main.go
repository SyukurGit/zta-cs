package main

import (
	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/config"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/internal/handler"
	"github.com/syukurgit/zta/internal/middleware"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/internal/service"
)

func main() {
	config.ConnectDB()
	// config.MigrateDB() // Uncomment sekali saja saat deployment awal untuk update struktur tabel

	// --- SETUP LAYERS ---

	// 1. AUDIT LAYER (Foundation)
	// Inisialisasi duluan karena dibutuhkan oleh TicketService & VerificationService
	auditRepo := repository.NewAuditRepository(config.DB)
	auditService := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	// 2. AUTH LAYER
	authHandler := &handler.AuthHandler{DB: config.DB}

	// 3. TICKET LAYER
	ticketRepo := repository.NewTicketRepository(config.DB)
	// Perhatikan: Kita inject auditService ke sini
	ticketService := service.NewTicketService(ticketRepo, auditService)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// 4. VERIFICATION LAYER
	verifRepo := repository.NewVerificationRepository(config.DB)
	// Perhatikan: Kita inject auditService ke sini juga
	verifService := service.NewVerificationService(verifRepo, auditService)
	verifHandler := handler.NewVerificationHandler(verifService)

	// 5. CHAT LAYER
	chatRepo := repository.NewChatRepository(config.DB)
	// Chat butuh ticketRepo untuk validasi kepemilikan tiket
	chatService := service.NewChatService(chatRepo, ticketRepo)
	chatHandler := handler.NewChatHandler(chatService)

	// --- SETUP ROUTER ---
	r := gin.Default()

	// Public Route
	r.POST("/login", authHandler.Login)

	// Verification Routes (Public but Secure via Token)
	r.GET("/verify/:token", verifHandler.GetVerificationPage)
	r.POST("/verify/:token", verifHandler.SubmitVerification)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// GROUP: USER
		userGroup := api.Group("/user")
		userGroup.Use(middleware.EnforceRole(domain.RoleUser))
		{
			userGroup.POST("/tickets", ticketHandler.CreateTicket)
			// Fitur Chat User
			userGroup.POST("/tickets/:id/chat", chatHandler.SendChat)
			userGroup.GET("/tickets/:id/chat", chatHandler.GetHistory)
		}

		// GROUP: CS
		csGroup := api.Group("/cs")
		csGroup.Use(middleware.EnforceRole(domain.RoleCS))
		{
			csGroup.GET("/tickets/open", ticketHandler.GetOpenTickets)
			csGroup.POST("/tickets/:id/claim", ticketHandler.ClaimTicket) // Limit 1 Tiket ada di Service
			csGroup.POST("/tickets/:id/start-verification", verifHandler.StartVerification)
			csGroup.POST("/tickets/:id/reset-password", ticketHandler.ResetPasswordAction)
			// Fitur Chat CS
			csGroup.POST("/tickets/:id/chat", chatHandler.SendChat)
			csGroup.GET("/tickets/:id/chat", chatHandler.GetHistory)
		}

		// GROUP: AUDITOR (NEW)
		auditorGroup := api.Group("/auditor")
		auditorGroup.Use(middleware.EnforceRole(domain.RoleAuditor))
		{
			auditorGroup.GET("/logs", auditHandler.GetLogs)
		}
	}

	r.Run(":8080")
}