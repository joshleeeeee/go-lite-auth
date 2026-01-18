package router

import (
	"github.com/gin-gonic/gin"
	"github.com/joshleeeeee/LiteSSO/internal/handler"
	"github.com/joshleeeeee/LiteSSO/internal/middleware"
)

// Setup initializes and returns the Gin router
func Setup(mode string) *gin.Engine {
	gin.SetMode(mode)

	r := gin.New()

	// Global middleware
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize handlers
	authHandler := handler.NewAuthHandler()

	// API routes
	api := r.Group("/api")
	{
		// Public auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/validate", authHandler.ValidateToken)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/user/info", authHandler.GetUserInfo)
		}
	}

	return r
}
