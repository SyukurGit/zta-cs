package main

import (
	"github.com/gin-contrib/cors"
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
	auditRepo := repository.NewAuditRepository(config.DB)
	auditService := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	// 2. AUTH LAYER
	authHandler := &handler.AuthHandler{DB: config.DB}

	// 3. TICKET LAYER
	ticketRepo := repository.NewTicketRepository(config.DB)
	ticketService := service.NewTicketService(ticketRepo, auditService)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// 4. VERIFICATION LAYER
	verifRepo := repository.NewVerificationRepository(config.DB)
	verifService := service.NewVerificationService(verifRepo, auditService)
	verifHandler := handler.NewVerificationHandler(verifService)

	// 5. CHAT LAYER
	chatRepo := repository.NewChatRepository(config.DB)
	chatService := service.NewChatService(chatRepo, ticketRepo)
	chatHandler := handler.NewChatHandler(chatService)

	// --- SETUP ROUTER ---
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public Route
	r.POST("/login", authHandler.Login)

	// Verification Routes (Public but Secure via Token)
	r.GET("/verify/:token", verifHandler.GetVerificationPage)
	r.POST("/verify/:token", verifHandler.SubmitVerification)
	r.POST("/reset-password", ticketHandler.SubmitUserResetPassword)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Endpoint Log Box untuk CS (Real-time monitoring)
		api.GET("/audit/tickets/:id", auditHandler.GetLogsByTicket)

		// GROUP: USER
		userGroup := api.Group("/user")
		userGroup.Use(middleware.EnforceRole(domain.RoleUser))
		{
			userGroup.POST("/tickets", ticketHandler.CreateTicket)
			userGroup.POST("/tickets/:id/chat", chatHandler.SendChat)
			userGroup.GET("/tickets/:id/chat", chatHandler.GetHistory)
			userGroup.POST("/tickets/:id/close", ticketHandler.CloseTicket)
			userGroup.GET("/tickets", ticketHandler.GetUserTickets)
			userGroup.GET("/tickets/:id", ticketHandler.GetTicketDetail)
		}

		// GROUP: CS
		csGroup := api.Group("/cs")
		csGroup.Use(middleware.EnforceRole(domain.RoleCS))
		{
			csGroup.GET("/tickets/open", ticketHandler.GetOpenTickets)
			csGroup.POST("/tickets/:id/claim", ticketHandler.ClaimTicket)
			csGroup.POST("/tickets/:id/start-verification", verifHandler.StartVerification)
			csGroup.POST("/tickets/:id/reset-password", ticketHandler.ResetPasswordAction)
			csGroup.GET("/tickets/history", ticketHandler.GetCSHistory)
			csGroup.POST("/tickets/:id/chat", chatHandler.SendChat)
			csGroup.GET("/tickets/:id/chat", chatHandler.GetHistory)
			csGroup.POST("/tickets/:id/close", ticketHandler.CloseTicket)
			csGroup.GET("/tickets/mine", ticketHandler.GetCSActiveTickets)
			csGroup.GET("/tickets/:id", ticketHandler.GetTicketDetail)
		}

		// GROUP: AUDITOR (Updated with Zero Trust Report Routes)
		auditorGroup := api.Group("/auditor")
		auditorGroup.Use(middleware.EnforceRole(domain.RoleAuditor))
		{
			auditorGroup.GET("/logs", auditHandler.GetLogs)                      // Log mentah (Immutable)
			auditorGroup.GET("/reports", auditHandler.GetAuditReports)           // Daftar laporan per tiket
			auditorGroup.GET("/tickets/:id/logs", auditHandler.GetLogsByTicket)  // Timeline detail log per tiket
			auditorGroup.GET("/tickets/:id/chat", chatHandler.GetHistory)       // Riwayat chat untuk audit
		}
	}

	r.Run(":8080")
}