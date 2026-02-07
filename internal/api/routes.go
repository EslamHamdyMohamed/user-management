package api

import (
	"user-management/internal/handler"
	"user-management/internal/middleware"

	"github.com/gin-contrib/pprof"
)

func (s *Server) SetupRoutes(authHandler *handler.AuthHandler, userHandler *handler.UserHandler) {
	// Register pprof routes
	if s.cfg.App.Environment != "production" {
		pprof.Register(s.router)
	}

	api := s.router.Group("/api/v1")

	// Health check
	api.GET("/health", s.healthCheck)

	// Public routes
	public := api.Group("/auth")
	{
		public.POST("/signin", authHandler.SignIn)
		public.POST("/signup", authHandler.Signup)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.Auth(s.jwtManager))
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
		}
	}
}
