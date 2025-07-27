package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PinRequest struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	CID          string          `gorm:"size:64;index;not null" json:"cid"`
	Status       string          `gorm:"size:20;default:'pending'" json:"status"`
	SizeBytes    int64           `gorm:"default:0" json:"size_bytes"`
	PriceFIL     decimal.Decimal `gorm:"type:decimal(18,8);default:0" json:"price_fil"`
	DurationDays int             `gorm:"not null" json:"duration_days"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FilecoinDeals []FilecoinDeal `gorm:"foreignKey:PinRequestID" json:"filecoin_deals,omitempty"`
}

func (PinRequest) TableName() string {
	return "pin_requests"
}

// Status constants
const (
	PinStatusPending   = "pending"
	PinStatusPinned    = "pinned"
	PinStatusFailed    = "failed"
	PinStatusCancelled = "cancelled"
)

// IsActive returns true if the pin request is in an active state
func (p *PinRequest) IsActive() bool {
	return p.Status == PinStatusPending || p.Status == PinStatusPinned
}

// CanBeCancelled returns true if the pin request can be cancelled
func (p *PinRequest) CanBeCancelled() bool {
	return p.Status == PinStatusPending
}
