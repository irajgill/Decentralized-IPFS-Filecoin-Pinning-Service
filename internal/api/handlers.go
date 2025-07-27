package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"pinning-service/internal/models"
	"pinning-service/internal/services"
)

type Handlers struct {
	dealService    *services.DealService
	pricingService *services.PricingService
	userService    *services.UserService
	logger         *logrus.Logger
}

type PinRequest struct {
	CID          string `json:"cid" binding:"required"`
	DurationDays int    `json:"duration_days" binding:"required,min=1"`
}

type PinResponse struct {
	ID           string  `json:"id"`
	CID          string  `json:"cid"`
	Status       string  `json:"status"`
	SizeBytes    int64   `json:"size_bytes"`
	PriceFIL     float64 `json:"price_fil"`
	DurationDays int     `json:"duration_days"`
	CreatedAt    string  `json:"created_at"`
}

func NewHandlers(dealService *services.DealService, pricingService *services.PricingService, userService *services.UserService, logger *logrus.Logger) *Handlers {
	return &Handlers{
		dealService:    dealService,
		pricingService: pricingService,
		userService:    userService,
		logger:         logger,
	}
}

// PostPin handles pin requests
func (h *Handlers) PostPin(c *gin.Context) {
	var req PinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	pinRequest := &models.PinRequest{
		ID:           uuid.New(),
		UserID:       userUUID,
		CID:          req.CID,
		DurationDays: req.DurationDays,
		Status:       "pending",
	}

	if err := h.dealService.SubmitPinRequest(c.Request.Context(), pinRequest); err != nil {
		h.logger.WithError(err).Error("Failed to submit pin request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit pin request"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"id":      pinRequest.ID.String(),
		"message": "Pin request submitted successfully",
	})
}

// GetPin retrieves pin request status
func (h *Handlers) GetPin(c *gin.Context) {
	pinID := c.Param("id")
	if pinID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pin ID is required"})
		return
	}

	pinUUID, err := uuid.Parse(pinID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pin ID format"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	pinRequest, err := h.dealService.GetPinRequest(c.Request.Context(), pinUUID, userID.(string))
	if err != nil {
		if err.Error() == "pin request not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pin request not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to get pin request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pin request"})
		return
	}

	response := PinResponse{
		ID:           pinRequest.ID.String(),
		CID:          pinRequest.CID,
		Status:       pinRequest.Status,
		SizeBytes:    pinRequest.SizeBytes,
		PriceFIL:     pinRequest.PriceFIL,
		DurationDays: pinRequest.DurationDays,
		CreatedAt:    pinRequest.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPins lists user's pin requests
func (h *Handlers) GetPins(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Parse query parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	status := c.Query("status")

	pins, total, err := h.dealService.GetUserPinRequests(c.Request.Context(), userID.(string), page, limit, status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user pin requests")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pin requests"})
		return
	}

	var responses []PinResponse
	for _, pin := range pins {
		responses = append(responses, PinResponse{
			ID:           pin.ID.String(),
			CID:          pin.CID,
			Status:       pin.Status,
			SizeBytes:    pin.SizeBytes,
			PriceFIL:     pin.PriceFIL,
			DurationDays: pin.DurationDays,
			CreatedAt:    pin.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"pins":  responses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// DeletePin cancels/unpins content
func (h *Handlers) DeletePin(c *gin.Context) {
	pinID := c.Param("id")
	if pinID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pin ID is required"})
		return
	}

	pinUUID, err := uuid.Parse(pinID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pin ID format"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	if err := h.dealService.CancelPinRequest(c.Request.Context(), pinUUID, userID.(string)); err != nil {
		if err.Error() == "pin request not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pin request not found"})
			return
		}
		h.logger.WithError(err).Error("Failed to cancel pin request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel pin request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pin request cancelled successfully"})
}

// GetDeals returns Filecoin deals for content
func (h *Handlers) GetDeals(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	deals, err := h.dealService.GetDealsForCID(c.Request.Context(), cid, userID.(string))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get deals for CID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deals": deals})
}

// PostRenewDeal renews expiring deals
func (h *Handlers) PostRenewDeal(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID is required"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	if err := h.dealService.RenewDealsForCID(c.Request.Context(), cid, userID.(string)); err != nil {
		h.logger.WithError(err).Error("Failed to renew deals for CID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew deals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deal renewal initiated"})
}

// GetPricing returns current pricing information
func (h *Handlers) GetPricing(c *gin.Context) {
	sizeBytes := int64(1024 * 1024 * 1024) // Default 1GB
	if s := c.Query("size_bytes"); s != "" {
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil {
			sizeBytes = parsed
		}
	}

	durationDays := 30 // Default 30 days
	if d := c.Query("duration_days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			durationDays = parsed
		}
	}

	price := h.pricingService.CalculatePrice(sizeBytes, durationDays)

	c.JSON(http.StatusOK, gin.H{
		"size_bytes":    sizeBytes,
		"duration_days": durationDays,
		"price_fil":     price,
	})
}

// GetMiners returns list of available storage miners
func (h *Handlers) GetMiners(c *gin.Context) {
	miners, err := h.dealService.GetAvailableMiners(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get available miners")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get miners"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"miners": miners})
}

// GetStats returns service statistics and health
func (h *Handlers) GetStats(c *gin.Context) {
	stats, err := h.dealService.GetServiceStats(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to get service stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HealthCheck returns health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": c.Request.Context().Value("timestamp"),
	})
}
