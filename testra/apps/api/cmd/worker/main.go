package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/analytics"
	"github.com/testra/testra/apps/api/internal/billing"
	"github.com/testra/testra/apps/api/internal/integrationhub"
	"github.com/testra/testra/apps/api/internal/intelligence"
	"github.com/testra/testra/apps/api/internal/notification"
	"github.com/testra/testra/apps/api/internal/queue"
	"github.com/testra/testra/apps/api/internal/shared/config"
	"github.com/testra/testra/apps/api/internal/shared/db"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	smtpCfg := notification.SMTPConfig{
		Host: cfg.SMTPHost,
		Port: cfg.SMTPPort,
		From: cfg.SMTPFrom,
	}

	notificationModule := notification.NewModule(database, smtpCfg.Host, smtpCfg.Port, smtpCfg.From)
	analyticsModule := analytics.New(database)
	intelligenceModule := intelligence.New(database, cfg.MLServiceURL)
	integrationhubModule := integrationhub.New(database)
	billingModule := billing.New(database, cfg.StripeSecretKey)

	runner := &Runner{
		db:             database,
		pollInterval:   getPollInterval(),
		notification:   notificationModule,
		analytics:      analyticsModule,
		intelligence:   intelligenceModule,
		integrationhub: integrationhubModule,
		billing:        billingModule,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Testra background worker started")
	runner.Run(ctx)
	log.Println("Testra background worker stopped")
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

type Runner struct {
	db             *sql.DB
	pollInterval   time.Duration
	notification   *notification.Module
	analytics      *analytics.Module
	intelligence   *intelligence.Module
	integrationhub *integrationhub.Module
	billing        *billing.Module
}

func (r *Runner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	// Run an initial pass immediately.
	r.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *Runner) processBatch(ctx context.Context) {
	for {
		done, err := r.processOne(ctx)
		if err != nil {
			log.Printf("worker error: %v", err)
		}
		if done {
			return
		}
	}
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

	if err := r.processJob(workCtx, job); err != nil {
		if markErr := queue.MarkFailed(ctx, tx, job.ID, job.Attempts+1, job.MaxAttempts, err.Error()); markErr != nil {
			return false, fmt.Errorf("mark failed %s: %w", job.ID, markErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return false, fmt.Errorf("commit failure state: %w", commitErr)
		}
		tx = nil
		return false, fmt.Errorf("job %s failed: %w", job.ID, err)
	}

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
