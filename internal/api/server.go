package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-management/internal/config"
	"user-management/internal/handler"
	"user-management/internal/middleware"
	"user-management/internal/repository"
	"user-management/internal/service"
	"user-management/internal/utils"
	"user-management/pkg/database"
	"user-management/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	cfg        *config.Config
	router     *gin.Engine
	server     *http.Server
	db         *database.Database
	jwtManager *utils.JWTManager
	logger     *logger.Logger
}

func NewServer(cfg *config.Config) *Server {
	// Initialize logger
	logger := logger.NewLogger(cfg.App.LogLevel, cfg.App.Environment)

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(
		middleware.Recovery(logger),
		middleware.Logging(logger),
		middleware.CORS(cfg.Server.CORS),
		gin.Recovery(),
	)

	return &Server{
		cfg:    cfg,
		router: router,
		logger: logger,
	}
}

func (s *Server) Setup() error {
	// Initialize database
	db, err := database.NewDatabase(&s.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	s.db = db

	// Run migrations
	if err := s.db.RunMigrations("internal/migration/migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize JWT manager
	s.jwtManager = utils.NewJWTManager(&s.cfg.JWT)
	passwordManager := utils.NewPasswordManager(bcrypt.DefaultCost)

	// Initialize repository
	userRepo := repository.NewUserRepository(s.db.DB)

	// Initialize services
	userService := service.NewUserService(userRepo, s.jwtManager, passwordManager)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, s.jwtManager)
	userHandler := handler.NewUserHandler(userService)

	// Setup routes
	s.SetupRoutes(authHandler, userHandler)

	// Create HTTP server with timeouts
	s.server = &http.Server{
		Addr:         ":" + s.cfg.Server.Port,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	return nil
}

func (s *Server) healthCheck(c *gin.Context) {
	// Check database health
	if err := s.db.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   s.cfg.App.Version,
	})
}

func (s *Server) Run() error {
	// Start server in goroutine
	go func() {
		s.logger.Info().Str("port", s.cfg.Server.Port).Msg("🚀 Server starting")
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info().Msg("Shutdown signal received")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Server forced to shutdown")
		return err
	}

	// Close database connection
	if err := s.db.Close(); err != nil {
		s.logger.Error().Err(err).Msg("Failed to close database connection")
	}

	s.logger.Info().Msg(" Server shutdown completed")
	return nil
}
