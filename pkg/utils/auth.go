package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"pinning-service/internal/storage"
)

var jwtSecret []byte

func InitJWT(secret string) {
	jwtSecret = []byte(secret)
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID string, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT validates a JWT token and returns the user ID
func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("invalid user_id claim")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}

// ValidateAPIKey validates an API key and returns the user ID
func ValidateAPIKey(db interface{}, apiKey string) (string, error) {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return "", fmt.Errorf("invalid database instance")
	}

	userRepo := storage.NewUserRepository(gormDB)
	user, err := userRepo.GetByAPIKey(context.Background(), apiKey)
	if err != nil {
		return "", fmt.Errorf("invalid API key")
	}

	return user.ID.String(), nil
}

// GenerateAPIKey generates a new API key
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
