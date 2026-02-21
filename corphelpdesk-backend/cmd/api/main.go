package main

import (
	"github.com/yourcompany/corphelpdesk-backend/internal/config"
	"github.com/yourcompany/corphelpdesk-backend/internal/handler/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/handler/tickets"
	"github.com/yourcompany/corphelpdesk-backend/internal/handler/messages"
	"github.com/yourcompany/corphelpdesk-backend/internal/handler/admin"
	"github.com/yourcompany/corphelpdesk-backend/internal/middleware"
	"github.com/yourcompany/corphelpdesk-backend/internal/repository/postgres"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/auth"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/tickets"
	"github.com/yourcompany/corphelpdesk-backend/internal/service/encryption"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize encryption service
	encService, err := encryption.NewAESEncryption(cfg.EncryptionKey)
	if err != nil {
		panic(err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	ticketRepo := postgres.NewTicketRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	// Initialize services
	authService := authservice.NewAuthService(userRepo, encService)
	ticketService := ticketservice.NewTicketService(ticketRepo, messageRepo)
	
	// Initialize handlers
	authHandler := authhandler.NewAuthHandler(authService)
	ticketHandler := tickethandler.NewTicketHandler(ticketService, authService)
	messageHandler := messagehandler.NewMessageHandler(ticketService, authService)
	adminHandler := adminhandler.NewAdminHandler(userRepo, ticketRepo, authService)

	// Setup Gin router
	r := gin.Default()

	// Global middleware
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/auth/verify", authHandler.VerifyTelegramAuth)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.TelegramAuthMiddleware())
		{
			protected.GET("/users/me", authHandler.GetProfile)
			protected.POST("/users/profile", authHandler.SubmitProfile)

			// Ticket routes
			protected.GET("/tickets", ticketHandler.GetTickets)
			protected.POST("/tickets", ticketHandler.CreateTicket)
			protected.GET("/tickets/:id", ticketHandler.GetTicketByID)
			protected.PATCH("/tickets/:id/status", middleware.RoleMiddleware("agent", "admin"), ticketHandler.UpdateTicketStatus)
			protected.PATCH("/tickets/:id/assign", middleware.RoleMiddleware("admin"), ticketHandler.AssignTicket)
			
			// Message routes
			protected.POST("/tickets/:id/messages", messageHandler.AddMessage)
		}

		// Admin routes
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.TelegramAuthMiddleware())
		adminRoutes.Use(middleware.RoleMiddleware("admin", "security_officer"))
		{
			adminRoutes.GET("/verification-requests", adminHandler.GetVerificationRequests)
			adminRoutes.POST("/verification/:user_id/decide", adminHandler.DecideVerification)
			adminRoutes.GET("/users", adminHandler.GetUsers)
			adminRoutes.POST("/users/:id/revoke", adminHandler.RevokeUser)
			
			// Security officer routes
			securityRoutes := adminRoutes.Group("/")
			securityRoutes.Use(middleware.RoleMiddleware("security_officer"))
			{
				securityRoutes.GET("/logs", adminHandler.GetAccessLogs)
			}
		}
	}

	// Start server
	r.Run(":" + cfg.Port)
}