package models

import (
	"time"

	"math/rand"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	APIKey    string          `gorm:"size:64;uniqueIndex;not null" json:"-"`
	Email     string          `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Balance   decimal.Decimal `gorm:"type:decimal(18,8);default:0" json:"balance"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	PinRequests []PinRequest `gorm:"foreignKey:UserID" json:"pin_requests,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate generates API key before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.APIKey == "" {
		u.APIKey = generateAPIKey()
	}
	return nil
}

func generateAPIKey() string {
	// Generate a 64-character API key
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	key := make([]byte, 64)
	for i := range key {
		key[i] = charset[rand.Intn(len(charset))]
	}
	return string(key)
}
