package main

import (
	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/config"
	"github.com/syukurgit/zta/internal/handler"
	"github.com/syukurgit/zta/internal/domain" // Pastikan import ini ada
	"github.com/syukurgit/zta/internal/middleware"
	"github.com/syukurgit/zta/internal/repository"
	"github.com/syukurgit/zta/internal/service"
)

func main() {
	config.ConnectDB()
	// config.MigrateDB()

	// --- SETUP LAYERS ---
	authHandler := &handler.AuthHandler{DB: config.DB}
	
	ticketRepo := repository.NewTicketRepository(config.DB)
	ticketService := service.NewTicketService(ticketRepo)
	ticketHandler := handler.NewTicketHandler(ticketService)

	// Initialize VerificationService and Handler
	verifRepo := repository.NewVerificationRepository(config.DB)
    verifService := service.NewVerificationService(verifRepo)
    verifHandler := handler.NewVerificationHandler(verifService)
	// --- SETUP ROUTER ---
	r := gin.Default()
	
	r.POST("/login", authHandler.Login)

	// --- PUBLIC ROUTES (Verification) ---
	// Tidak butuh AuthMiddleware karena tokennya ada di URL (Session ID)
	r.GET("/verify/:token", verifHandler.GetVerificationPage)
	r.POST("/verify/:token", verifHandler.SubmitVerification)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) 
	{
		// GROUP: USER
		userGroup := api.Group("/user")
		// PERBAIKAN: Gunakan domain.RoleUser (sesuai yang kita buat tadi)
		userGroup.Use(middleware.EnforceRole(domain.RoleUser)) 
		{
			userGroup.POST("/tickets", ticketHandler.CreateTicket)
		}

		// GROUP: CS
		csGroup := api.Group("/cs")
		// PERBAIKAN: Gunakan domain.RoleCS
		csGroup.Use(middleware.EnforceRole(domain.RoleCS))
		{
			csGroup.GET("/tickets/open", ticketHandler.GetOpenTickets)
			csGroup.POST("/tickets/:id/claim", ticketHandler.ClaimTicket)
			csGroup.POST("/tickets/:id/start-verification", verifHandler.StartVerification)
		}
	}

	r.Run(":8080")
}