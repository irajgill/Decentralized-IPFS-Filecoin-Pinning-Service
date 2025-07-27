package workers

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gocraft/work"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"pinning-service/internal/filecoin"
	"pinning-service/internal/ipfs"
	"pinning-service/internal/services"
	"pinning-service/internal/storage"
	"pinning-service/pkg/config"
)

type WorkerPool struct {
	pool   *work.WorkerPool
	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Logger
}

func NewWorkerPool(ctx context.Context, db *gorm.DB, redisClient *redis.Client, cfg *config.Config, logger *logrus.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	// Initialize clients
	ipfsClient := ipfs.NewClient(cfg.IPFS.APIURL, cfg.IPFS.Timeout)
	lotusClient, err := filecoin.NewLotusClient(cfg.Filecoin.LotusAPI, cfg.Filecoin.LotusToken)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Lotus client")
	}

	// Initialize repositories
	pinRepo := storage.NewPinRequestRepository(db)
	dealRepo := storage.NewFilecoinDealRepository(db)

	// Initialize services
	pricingService := services.NewPricingService(cfg)
	dealService := services.NewDealService(ipfsClient, lotusClient, pinRepo, dealRepo, pricingService, redisClient, cfg, logger)

	// Create job context
	jobCtx := &JobContext{
		DealService: dealService,
		Logger:      logger,
	}

	// Create worker pool
	pool := work.NewWorkerPool(jobCtx, cfg.Workers.Concurrency, cfg.Redis.Namespace, redisClient.Pool())

	// Add middleware
	pool.Middleware((*JobContext).LogMiddleware)
	pool.Middleware((*JobContext).ErrorMiddleware)

	// Register job handlers
	pool.Job("process_pin", (*JobContext).ProcessPin)
	pool.Job("monitor_deals", (*JobContext).MonitorDeals)
	pool.Job("renew_expiring", (*JobContext).RenewExpiring)
	pool.Job("cleanup_failed", (*JobContext).CleanupFailed)

	return &WorkerPool{
		pool:   pool,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
}

func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool")
	wp.pool.Start()

	// Schedule periodic jobs
	go wp.schedulePeriodicJobs()
}

func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool")
	wp.cancel()
	wp.pool.Stop()
}

func (wp *WorkerPool) schedulePeriodicJobs() {
	// Monitor deals every 5 minutes
	dealMonitorTicker := time.NewTicker(5 * time.Minute)
	defer dealMonitorTicker.Stop()

	// Check for expiring deals every hour
	renewalTicker := time.NewTicker(1 * time.Hour)
	defer renewalTicker.Stop()

	// Cleanup failed requests every 6 hours
	cleanupTicker := time.NewTicker(6 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-dealMonitorTicker.C:
			wp.enqueueJob("monitor_deals", nil)
		case <-renewalTicker.C:
			wp.enqueueJob("renew_expiring", nil)
		case <-cleanupTicker.C:
			wp.enqueueJob("cleanup_failed", nil)
		}
	}
}

func (wp *WorkerPool) enqueueJob(jobName string, args map[string]interface{}) {
	_, err := wp.pool.Enqueue(jobName, args)
	if err != nil {
		wp.logger.WithError(err).WithField("job", jobName).Error("Failed to enqueue job")
	}
}

// Middleware functions

func (c *JobContext) LogMiddleware(job *work.Job, next work.NextMiddlewareFunc) error {
	start := time.Now()

	c.Logger.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"job_name": job.Name,
	}).Info("Job started")

	err := next()

	duration := time.Since(start)

	fields := logrus.Fields{
		"job_id":   job.ID,
		"job_name": job.Name,
		"duration": duration,
	}

	if err != nil {
		c.Logger.WithError(err).WithFields(fields).Error("Job failed")
	} else {
		c.Logger.WithFields(fields).Info("Job completed")
	}

	return err
}

func (c *JobContext) ErrorMiddleware(job *work.Job, next work.NextMiddlewareFunc) error {
	err := next()
	if err != nil {
		// Could implement retry logic here
		c.Logger.WithError(err).WithField("job_id", job.ID).Error("Job error")
	}
	return err
}
