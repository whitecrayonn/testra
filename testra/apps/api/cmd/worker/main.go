package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/analytics"
	"github.com/testra/testra/apps/api/internal/audit"
	"github.com/testra/testra/apps/api/internal/billing"
	"github.com/testra/testra/apps/api/internal/integrationhub"
	"github.com/testra/testra/apps/api/internal/intelligence"
	"github.com/testra/testra/apps/api/internal/metrics"
	"github.com/testra/testra/apps/api/internal/notification"
	"github.com/testra/testra/apps/api/internal/queue"
	"github.com/testra/testra/apps/api/internal/shared/config"
	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/eventbus"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	smtpCfg := notification.SMTPConfig{
		Host:           cfg.SMTPHost,
		Port:           cfg.SMTPPort,
		From:           cfg.SMTPFrom,
		Username:       cfg.SMTPUsername,
		SecretProvider: cfg.SecretProvider(),
		PasswordSecret: cfg.SMTPPasswordSecret,
	}

	notificationModule := notification.NewModule(database, smtpCfg.Host, smtpCfg.Port, smtpCfg.From, smtpCfg.Username, smtpCfg.SecretProvider, smtpCfg.PasswordSecret)
	analyticsModule := analytics.New(database)
	intelligenceModule := intelligence.New(database, cfg.MLServiceURL)
	dbHandle := db.Wrap(database)
	auditSvc := audit.NewModule(dbHandle).Service
	eventBus := eventbus.New(256)
	integrationhubModule := integrationhub.New(database, auditSvc, eventBus)
	billingModule := billing.New(database, cfg.StripeSecretKey)

	runner := &Runner{
		db:              database,
		pollInterval:    getPollInterval(),
		cleanupInterval: getCleanupInterval(),
		jobRetention:    getJobRetention(),
		notification:    notificationModule,
		analytics:       analyticsModule,
		intelligence:    intelligenceModule,
		integrationhub:  integrationhubModule,
		billing:         billingModule,
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	addr := ":" + metricsPort
	log.Printf("worker metrics server listening on %s", addr)
	metricsSrv := &http.Server{
		Addr:              addr,
		Handler:           metrics.Handler(database),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("metrics server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Testra background worker started")
	runner.Run(ctx)
	log.Println("Testra background worker stopped")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("metrics server shutdown error: %v", err)
	}
}

func getPollInterval() time.Duration {
	s := os.Getenv("WORKER_POLL_INTERVAL_SECONDS")
	if s == "" {
		s = "5"
	}
	secs, err := strconv.Atoi(s)
	if err != nil || secs < 1 {
		secs = 5
	}
	return time.Duration(secs) * time.Second
}

func getCleanupInterval() time.Duration {
	s := os.Getenv("WORKER_CLEANUP_INTERVAL_SECONDS")
	if s == "" {
		s = "300"
	}
	secs, err := strconv.Atoi(s)
	if err != nil || secs < 1 {
		secs = 300
	}
	return time.Duration(secs) * time.Second
}

func getJobRetention() time.Duration {
	s := os.Getenv("WORKER_JOB_RETENTION_HOURS")
	if s == "" {
		s = "24"
	}
	hours, err := strconv.Atoi(s)
	if err != nil || hours < 1 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}

type Runner struct {
	db              *sql.DB
	pollInterval    time.Duration
	cleanupInterval time.Duration
	jobRetention    time.Duration
	notification    *notification.Module
	analytics       *analytics.Module
	intelligence    *intelligence.Module
	integrationhub  *integrationhub.Module
	billing         *billing.Module
}

func (r *Runner) Run(ctx context.Context) {
	cleanupInterval := r.cleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute
	}
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	nextPoll := r.pollInterval
	if nextPoll <= 0 {
		nextPoll = 5 * time.Second
	}
	maxBackoff := 60 * time.Second
	if r.pollInterval > maxBackoff {
		maxBackoff = r.pollInterval
	}

	// Run an initial pass immediately.
	worked := r.processBatch(ctx)
	if worked {
		nextPoll = r.pollInterval
	}

	timer := time.NewTimer(nextPoll)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cleanupTicker.C:
			r.cleanup(ctx)
			continue
		case <-timer.C:
			worked := r.processBatch(ctx)
			if worked {
				nextPoll = r.pollInterval
			} else {
				nextPoll *= 2
				if nextPoll > maxBackoff {
					nextPoll = maxBackoff
				}
			}
			timer.Reset(nextPoll)
		}
	}
}

func (r *Runner) cleanup(ctx context.Context) {
	retention := r.jobRetention
	if retention <= 0 {
		retention = 24 * time.Hour
	}
	deleted, err := queue.DeleteOldCompleted(ctx, r.db, retention)
	if err != nil {
		log.Printf("worker cleanup error: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("worker cleanup: removed %d terminal jobs", deleted)
	}
}

func (r *Runner) processBatch(ctx context.Context) bool {
	worked := false
	for {
		done, err := r.processOne(ctx)
		if err != nil {
			log.Printf("worker error: %v", err)
			// Stop the batch and let Run apply exponential backoff. The
			// failed job is already rescheduled or dead-lettered inside
			// processOne; repeated DB errors should not spin CPU.
			return worked
		}
		if done {
			break
		}
		worked = true
	}
	return worked
}

func (r *Runner) processOne(ctx context.Context) (bool, error) {
	tx, job, err := queue.DequeueOne(ctx, r.db, "default")
	if err != nil {
		return false, err
	}
	if job == nil {
		return true, nil
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	// Set tenant context on the transaction for RLS.
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", job.TenantID.String()); err != nil {
		return false, fmt.Errorf("set tenant id: %w", err)
	}
	workCtx := db.WithTx(ctx, tx)
	workCtx = db.WithTenantID(workCtx, job.TenantID)

	start := time.Now()
	if err := r.processJob(workCtx, job); err != nil {
		status := "retry"
		if job.Attempts+1 >= job.MaxAttempts {
			status = "dead_letter"
		}
		metrics.RecordJob(job.JobType, status, time.Since(start))

		if markErr := queue.MarkFailed(ctx, tx, job.ID, job.Attempts+1, job.MaxAttempts, err.Error()); markErr != nil {
			return false, fmt.Errorf("mark failed %s: %w", job.ID, markErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return false, fmt.Errorf("commit failure state: %w", commitErr)
		}
		tx = nil
		return false, fmt.Errorf("job %s failed: %w", job.ID, err)
	}

	metrics.RecordJob(job.JobType, "success", time.Since(start))

	if err := queue.MarkDone(ctx, tx, job.ID); err != nil {
		return false, fmt.Errorf("mark done %s: %w", job.ID, err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit done state: %w", err)
	}
	tx = nil
	return false, nil
}

func (r *Runner) processJob(ctx context.Context, job *queue.Job) error {
	switch job.JobType {
	case "notification:send":
		var payload struct {
			OrganizationID string `json:"organization_id"`
			WorkspaceID    string `json:"workspace_id"`
			UserID         string `json:"user_id"`
			Type           string `json:"type"`
			Title          string `json:"title"`
			Body           string `json:"body"`
			Link           string `json:"link"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		orgID, _ := uuid.Parse(payload.OrganizationID)
		wsID, _ := uuid.Parse(payload.WorkspaceID)
		userID, _ := uuid.Parse(payload.UserID)
		_, err := r.notification.Service.Send(ctx, notification.SendInput{
			OrganizationID: orgID,
			WorkspaceID:    wsID,
			UserID:         userID,
			Type:           payload.Type,
			Title:          payload.Title,
			Body:           payload.Body,
			Link:           payload.Link,
		})
		return err

	case "analytics:aggregate":
		var payload struct {
			WorkspaceID string `json:"workspace_id"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		wsID, err := uuid.Parse(payload.WorkspaceID)
		if err != nil {
			return err
		}
		err = r.analytics.Service.AggregateMetrics(ctx, wsID, nil)
		return err

	case "intelligence:predict":
		var payload struct {
			WorkspaceID   string                         `json:"workspace_id"`
			TestCaseID    string                         `json:"test_case_id"`
			TestCaseTitle string                         `json:"test_case_title"`
			History       []intelligence.RunHistoryPoint `json:"history"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		wsID, err := uuid.Parse(payload.WorkspaceID)
		if err != nil {
			return err
		}
		_, err = r.intelligence.Service.PredictFlaky(ctx, intelligence.PredictFlakyInput{
			TestCaseID:    payload.TestCaseID,
			TestCaseTitle: payload.TestCaseTitle,
			History:       payload.History,
		}, wsID)
		return err

	case "integration:dispatch":
		var payload struct {
			WorkspaceID   string                 `json:"workspace_id"`
			IntegrationID string                 `json:"integration_id"`
			EventType     string                 `json:"event_type"`
			Payload       map[string]interface{} `json:"payload"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		wsID, err := uuid.Parse(payload.WorkspaceID)
		if err != nil {
			return err
		}
		intID, err := uuid.Parse(payload.IntegrationID)
		if err != nil {
			return err
		}
		_, err = r.integrationhub.Service.DispatchEvent(ctx, integrationhub.DispatchEventInput{
			WorkspaceID:   wsID,
			IntegrationID: intID,
			EventType:     payload.EventType,
			Payload:       payload.Payload,
		}, uuid.Nil)
		return err

	case "integration:retry":
		var payload struct {
			EventID string `json:"event_id"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		eventID, err := uuid.Parse(payload.EventID)
		if err != nil {
			return err
		}
		_, err = r.integrationhub.Service.RetryEvent(ctx, eventID, uuid.Nil)
		return err

	case "billing:sync":
		var payload struct {
			OrganizationID string `json:"organization_id"`
		}
		if err := job.ParsePayload(&payload); err != nil {
			return err
		}
		orgID, err := uuid.Parse(payload.OrganizationID)
		if err != nil {
			return err
		}
		_, err = r.billing.Service.GetSubscription(ctx, orgID)
		return err

	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}
