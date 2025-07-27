package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"pinning-service/pkg/config"
	"pinning-service/pkg/utils"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logger.WithFields(logrus.Fields{
				"status_code":   param.StatusCode,
				"latency":       param.Latency,
				"client_ip":     param.ClientIP,
				"method":        param.Method,
				"path":          param.Path,
				"user_agent":    param.Request.UserAgent(),
				"response_size": param.BodySize,
			}).Info("HTTP Request")
			return ""
		},
		Output: logger.Writer(),
	})
}

// RecoveryMiddleware handles panics
func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(logger.Writer(), func(c *gin.Context, recovered interface{}) {
		logger.WithField("panic", recovered).Error("Panic recovered")
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware validates JWT tokens or API keys
func AuthMiddleware(db interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check and pricing endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/pricing" {
			c.Next()
			return
		}

		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
			c.Abort()
			return
		}

		// Try JWT validation first
		userID, err := utils.ValidateJWT(token)
		if err == nil {
			c.Set("userID", userID)
			c.Set("authType", "jwt")
			c.Next()
			return
		}

		// Try API key validation
		userID, err = utils.ValidateAPIKey(db, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("authType", "api_key")
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(redisClient *redis.Client, cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		// Check current request count
		ctx := context.Background()
		current, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// If Redis is down, allow request
			c.Next()
			return
		}

		if current >= cfg.RateLimit.RequestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			// If Redis is down, allow request
			c.Next()
			return
		}

		c.Next()
	})
}

// extractToken extracts JWT token or API key from request
func extractToken(c *gin.Context) string {
	// Check Authorization header for Bearer token
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}

	// Check X-API-Key header
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Check query parameter
	return c.Query("token")
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.GenerateRequestID()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
