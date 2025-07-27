package models

import (
	"time"

	"github.com/google/uuid"
)

type FilecoinDeal struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PinRequestID  uuid.UUID `gorm:"type:uuid;index;not null" json:"pin_request_id"`
	DealCID       string    `gorm:"size:64;index" json:"deal_cid"`
	MinerID       string    `gorm:"size:20;not null" json:"miner_id"`
	StartEpoch    int64     `gorm:"not null" json:"start_epoch"`
	EndEpoch      int64     `gorm:"not null" json:"end_epoch"`
	Status        string    `gorm:"size:20;default:'pending'" json:"status"`
	StoragePrice  float64   `gorm:"type:decimal(18,8);default:0" json:"storage_price"`
	RetrievalCost float64   `gorm:"type:decimal(18,8);default:0" json:"retrieval_cost"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	PinRequest PinRequest `gorm:"foreignKey:PinRequestID" json:"pin_request,omitempty"`
}

func (FilecoinDeal) TableName() string {
	return "filecoin_deals"
}

// Status constants
const (
	DealStatusPending   = "pending"
	DealStatusPublished = "published"
	DealStatusActive    = "active"
	DealStatusExpired   = "expired"
	DealStatusSlashed   = "slashed"
	DealStatusFailed    = "failed"
	DealStatusCancelled = "cancelled"
)

// IsActive returns true if the deal is currently active
func (d *FilecoinDeal) IsActive() bool {
	return d.Status == DealStatusActive
}

// IsExpired returns true if the deal has expired
func (d *FilecoinDeal) IsExpired() bool {
	return d.Status == DealStatusExpired
}

// NeedsRenewal returns true if the deal is close to expiration
func (d *FilecoinDeal) NeedsRenewal(currentEpoch int64, renewalThreshold int64) bool {
	return d.IsActive() && (d.EndEpoch-currentEpoch) <= renewalThreshold
}
