package services

import (
	"pinning-service/pkg/config"
)

type PricingService struct {
	config *config.Config
}

func NewPricingService(config *config.Config) *PricingService {
	return &PricingService{
		config: config,
	}
}

// CalculatePrice calculates storage price in FIL
func (s *PricingService) CalculatePrice(sizeBytes int64, durationDays int) float64 {
	// Convert bytes to GB
	sizeGB := float64(sizeBytes) / (1024 * 1024 * 1024)

	// Convert days to months for pricing calculation
	durationMonths := float64(durationDays) / 30.0

	// Base price calculation
	basePrice := sizeGB * durationMonths * s.config.Pricing.BasePricePerGBPerMonth

	// Apply markup
	markup := basePrice * (s.config.Pricing.MarkupPercentage / 100.0)
	totalPrice := basePrice + markup

	// Ensure minimum price
	if sizeBytes < s.config.Pricing.MinimumDealSize {
		minPrice := float64(s.config.Pricing.MinimumDealSize) / (1024 * 1024 * 1024) *
			durationMonths * s.config.Pricing.BasePricePerGBPerMonth
		if totalPrice < minPrice {
			totalPrice = minPrice
		}
	}

	return totalPrice
}

// GetPricingInfo returns current pricing configuration
func (s *PricingService) GetPricingInfo() map[string]interface{} {
	return map[string]interface{}{
		"base_price_per_gb_per_month": s.config.Pricing.BasePricePerGBPerMonth,
		"markup_percentage":           s.config.Pricing.MarkupPercentage,
		"minimum_deal_size":           s.config.Pricing.MinimumDealSize,
		"currency":                    "FIL",
	}
}
