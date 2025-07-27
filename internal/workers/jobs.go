package workers

import (
	"context"
	"fmt"

	"github.com/gocraft/work"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"pinning-service/internal/services"
)

type JobContext struct {
	DealService *services.DealService
	Logger      *logrus.Logger
}

// ProcessPin processes a pin request job
func (c *JobContext) ProcessPin(job *work.Job) error {
	c.Logger.WithField("job_id", job.ID).Info("Processing pin job")

	// Extract pin ID from job arguments
	pinIDStr, ok := job.ArgString("pin_id")
	if !ok {
		return fmt.Errorf("missing pin_id argument")
	}

	pinID, err := uuid.Parse(pinIDStr)
	if err != nil {
		return fmt.Errorf("invalid pin_id format: %w", err)
	}

	// Process the pin request
	ctx := context.Background()
	if err := c.DealService.ProcessPinRequest(ctx, pinID); err != nil {
		c.Logger.WithError(err).WithField("pin_id", pinID).Error("Failed to process pin request")
		return err
	}

	c.Logger.WithField("pin_id", pinID).Info("Pin request processed successfully")
	return nil
}

// MonitorDeals monitors existing deals for status changes
func (c *JobContext) MonitorDeals(job *work.Job) error {
	c.Logger.Info("Monitoring deals")

	ctx := context.Background()
	if err := c.DealService.MonitorActiveDeals(ctx); err != nil {
		c.Logger.WithError(err).Error("Failed to monitor deals")
		return err
	}

	return nil
}

// RenewExpiring renews deals that are close to expiration
func (c *JobContext) RenewExpiring(job *work.Job) error {
	c.Logger.Info("Checking for expiring deals")

	ctx := context.Background()
	if err := c.DealService.RenewExpiringDeals(ctx); err != nil {
		c.Logger.WithError(err).Error("Failed to renew expiring deals")
		return err
	}

	return nil
}

// CleanupFailed cleans up failed pin requests
func (c *JobContext) CleanupFailed(job *work.Job) error {
	c.Logger.Info("Cleaning up failed requests")

	ctx := context.Background()
	if err := c.DealService.CleanupFailedRequests(ctx); err != nil {
		c.Logger.WithError(err).Error("Failed to cleanup failed requests")
		return err
	}

	return nil
}
