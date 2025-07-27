package api

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"pinning-service/internal/filecoin"
	"pinning-service/internal/ipfs"
	"pinning-service/internal/services"
	"pinning-service/internal/storage"
	"pinning-service/pkg/config"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config, logger *logrus.Logger) {
	// Initialize clients
	ipfsClient := ipfs.NewClient(cfg.IPFS.APIURL, cfg.IPFS.Timeout)
	lotusClient, err := filecoin.NewLotusClient(cfg.Filecoin.LotusAPI, cfg.Filecoin.LotusToken)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Lotus client")
	}

	// Initialize repositories
	userRepo := storage.NewUserRepository(db)
	pinRepo := storage.NewPinRequestRepository(db)
	dealRepo := storage.NewFilecoinDealRepository(db)

	// Initialize services
	pricingService := services.NewPricingService(cfg)
	userService := services.NewUserService(userRepo, cfg, logger)
	dealService := services.NewDealService(ipfsClient, lotusClient, pinRepo, dealRepo, pricingService, redisClient, cfg, logger)

	// Initialize handlers
	handlers := NewHandlers(dealService, pricingService, userService, logger)

	// Add auth middleware to all routes except health and pricing
	authGroup := router.Group("/")
	authGroup.Use(AuthMiddleware(db))

	// Core pin management endpoints
	authGroup.POST("/pin", handlers.PostPin)
	authGroup.GET("/pin/:id", handlers.GetPin)
	authGroup.GET("/pins", handlers.GetPins)
	authGroup.DELETE("/pin/:id", handlers.DeletePin)

	// Deal management endpoints
	authGroup.GET("/deals/:cid", handlers.GetDeals)
	authGroup.POST("/deals/:cid/renew", handlers.PostRenewDeal)

	// Public endpoints (no auth required)
	router.GET("/health", handlers.HealthCheck)
	router.GET("/pricing", handlers.GetPricing)
	router.GET("/miners", handlers.GetMiners)
	router.GET("/stats", handlers.GetStats)

	// API versioning
	v1 := router.Group("/api/v1")
	v1.Use(AuthMiddleware(db))
	{
		v1.POST("/pin", handlers.PostPin)
		v1.GET("/pin/:id", handlers.GetPin)
		v1.GET("/pins", handlers.GetPins)
		v1.DELETE("/pin/:id", handlers.DeletePin)
		v1.GET("/deals/:cid", handlers.GetDeals)
		v1.POST("/deals/:cid/renew", handlers.PostRenewDeal)
	}

	// Public v1 endpoints
	v1Public := router.Group("/api/v1")
	{
		v1Public.GET("/health", handlers.HealthCheck)
		v1Public.GET("/pricing", handlers.GetPricing)
		v1Public.GET("/miners", handlers.GetMiners)
		v1Public.GET("/stats", handlers.GetStats)
	}
}
