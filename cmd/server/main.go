package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pinning-service/internal/api"
	"pinning-service/internal/storage"
	"pinning-service/pkg/config"
	"pinning-service/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := utils.NewLogger(cfg)
	logger.Info("Starting IPFS-Filecoin Pinning Service")

	// Initialize database
	db, err := storage.InitPostgres(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}

	// Run migrations
	if err := storage.RunMigrations(cfg); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}

	// Initialize Redis
	redisClient, err := storage.InitRedis(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(
		api.LoggingMiddleware(logger),
		api.RecoveryMiddleware(logger),
		api.CORSMiddleware(cfg),
		api.RateLimitMiddleware(redisClient, cfg),
	)

	// Register routes
	api.RegisterRoutes(router, db, redisClient, cfg, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("address", srv.Addr).Info("HTTP server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}
