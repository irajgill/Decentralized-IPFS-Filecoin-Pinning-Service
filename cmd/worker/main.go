package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"pinning-service/internal/storage"
	"pinning-service/internal/workers"
	"pinning-service/pkg/config"
	"pinning-service/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := utils.NewLogger(cfg)
	logger.Info("Starting IPFS-Filecoin Pinning Service Worker")

	// Initialize database
	db, err := storage.InitPostgres(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}

	// Initialize Redis
	redisClient, err := storage.InitRedis(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}

	// Create worker context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize worker pool
	workerPool := workers.NewWorkerPool(ctx, db, redisClient, cfg, logger)

	// Start the worker pool
	workerPool.Start()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")

	// Stop the worker pool
	workerPool.Stop()

	logger.Info("Worker exited")
}
