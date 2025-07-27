package storage

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pinning-service/internal/models"
)

// UserRepository defines user data access methods
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PinRequestRepository defines pin request data access methods
type PinRequestRepository interface {
	Create(ctx context.Context, pinRequest *models.PinRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PinRequest, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int, status string) ([]*models.PinRequest, int64, error)
	GetByCID(ctx context.Context, cid string) ([]*models.PinRequest, error)
	Update(ctx context.Context, pinRequest *models.PinRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetPendingRequests(ctx context.Context, limit int) ([]*models.PinRequest, error)
}

// FilecoinDealRepository defines Filecoin deal data access methods
type FilecoinDealRepository interface {
	Create(ctx context.Context, deal *models.FilecoinDeal) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.FilecoinDeal, error)
	GetByPinRequestID(ctx context.Context, pinRequestID uuid.UUID) ([]*models.FilecoinDeal, error)
	GetByCID(ctx context.Context, cid string) ([]*models.FilecoinDeal, error)
	GetByMinerID(ctx context.Context, minerID string) ([]*models.FilecoinDeal, error)
	Update(ctx context.Context, deal *models.FilecoinDeal) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiringDeals(ctx context.Context, epochThreshold int64) ([]*models.FilecoinDeal, error)
	GetActiveDeals(ctx context.Context) ([]*models.FilecoinDeal, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "api_key = ?", apiKey).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

// pinRequestRepository implements PinRequestRepository
type pinRequestRepository struct {
	db *gorm.DB
}

func NewPinRequestRepository(db *gorm.DB) PinRequestRepository {
	return &pinRequestRepository{db: db}
}

func (r *pinRequestRepository) Create(ctx context.Context, pinRequest *models.PinRequest) error {
	return r.db.WithContext(ctx).Create(pinRequest).Error
}

func (r *pinRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PinRequest, error) {
	var pinRequest models.PinRequest
	err := r.db.WithContext(ctx).Preload("FilecoinDeals").First(&pinRequest, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pinRequest, nil
}

func (r *pinRequestRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int, status string) ([]*models.PinRequest, int64, error) {
	var pinRequests []*models.PinRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&models.PinRequest{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	query.Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Preload("FilecoinDeals").Offset(offset).Limit(limit).Order("created_at DESC").Find(&pinRequests).Error

	return pinRequests, total, err
}

func (r *pinRequestRepository) GetByCID(ctx context.Context, cid string) ([]*models.PinRequest, error) {
	var pinRequests []*models.PinRequest
	err := r.db.WithContext(ctx).Preload("FilecoinDeals").Find(&pinRequests, "cid = ?", cid).Error
	return pinRequests, err
}

func (r *pinRequestRepository) Update(ctx context.Context, pinRequest *models.PinRequest) error {
	return r.db.WithContext(ctx).Save(pinRequest).Error
}

func (r *pinRequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.PinRequest{}, "id = ?", id).Error
}

func (r *pinRequestRepository) GetPendingRequests(ctx context.Context, limit int) ([]*models.PinRequest, error) {
	var pinRequests []*models.PinRequest
	err := r.db.WithContext(ctx).Where("status = ?", models.PinStatusPending).Limit(limit).Find(&pinRequests).Error
	return pinRequests, err
}

// filecoinDealRepository implements FilecoinDealRepository
type filecoinDealRepository struct {
	db *gorm.DB
}

func NewFilecoinDealRepository(db *gorm.DB) FilecoinDealRepository {
	return &filecoinDealRepository{db: db}
}

func (r *filecoinDealRepository) Create(ctx context.Context, deal *models.FilecoinDeal) error {
	return r.db.WithContext(ctx).Create(deal).Error
}

func (r *filecoinDealRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.FilecoinDeal, error) {
	var deal models.FilecoinDeal
	err := r.db.WithContext(ctx).Preload("PinRequest").First(&deal, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *filecoinDealRepository) GetByPinRequestID(ctx context.Context, pinRequestID uuid.UUID) ([]*models.FilecoinDeal, error) {
	var deals []*models.FilecoinDeal
	err := r.db.WithContext(ctx).Find(&deals, "pin_request_id = ?", pinRequestID).Error
	return deals, err
}

func (r *filecoinDealRepository) GetByCID(ctx context.Context, cid string) ([]*models.FilecoinDeal, error) {
	var deals []*models.FilecoinDeal
	err := r.db.WithContext(ctx).
		Joins("JOIN pin_requests ON filecoin_deals.pin_request_id = pin_requests.id").
		Where("pin_requests.cid = ?", cid).
		Find(&deals).Error
	return deals, err
}

func (r *filecoinDealRepository) GetByMinerID(ctx context.Context, minerID string) ([]*models.FilecoinDeal, error) {
	var deals []*models.FilecoinDeal
	err := r.db.WithContext(ctx).Find(&deals, "miner_id = ?", minerID).Error
	return deals, err
}

func (r *filecoinDealRepository) Update(ctx context.Context, deal *models.FilecoinDeal) error {
	return r.db.WithContext(ctx).Save(deal).Error
}

func (r *filecoinDealRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.FilecoinDeal{}, "id = ?", id).Error
}

func (r *filecoinDealRepository) GetExpiringDeals(ctx context.Context, epochThreshold int64) ([]*models.FilecoinDeal, error) {
	var deals []*models.FilecoinDeal
	err := r.db.WithContext(ctx).
		Where("status = ? AND end_epoch <= ?", models.DealStatusActive, epochThreshold).
		Preload("PinRequest").
		Find(&deals).Error
	return deals, err
}

func (r *filecoinDealRepository) GetActiveDeals(ctx context.Context) ([]*models.FilecoinDeal, error) {
	var deals []*models.FilecoinDeal
	err := r.db.WithContext(ctx).Where("status = ?", models.DealStatusActive).Find(&deals).Error
	return deals, err
}
