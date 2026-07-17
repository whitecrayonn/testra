package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID
	QueueName   string
	JobType     string
	Payload     json.RawMessage
	Status      string
	Attempts    int
	MaxAttempts int
	TenantID    uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Enqueue inserts a new job into the queue.
func Enqueue(ctx context.Context, db *sql.DB, tenantID uuid.UUID, queueName, jobType string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire db connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", tenantID.String()); err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	_, err = conn.ExecContext(ctx,
		`INSERT INTO queue_jobs (queue_name, job_type, payload, status, tenant_id, scheduled_at, created_at, updated_at)
		 VALUES ($1, $2, $3, 'pending', $4, NOW(), NOW(), NOW())`,
		queueName, jobType, payloadJSON, tenantID)
	return err
}

// DequeueOne locks and returns the next pending job along with the transaction
// that owns the lock. The caller must commit or rollback the transaction.
func DequeueOne(ctx context.Context, db *sql.DB, queueName string) (*sql.Tx, *Job, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	row := tx.QueryRowContext(ctx,
		`SELECT id, queue_name, job_type, payload::text, status, attempts, max_attempts, tenant_id, created_at, updated_at
		 FROM queue_jobs
		 WHERE queue_name = $1 AND status = 'pending' AND scheduled_at <= NOW()
		 ORDER BY created_at ASC
		 FOR UPDATE SKIP LOCKED
		 LIMIT 1`,
		queueName)

	var job Job
	var payloadStr string
	if err := row.Scan(&job.ID, &job.QueueName, &job.JobType, &payloadStr, &job.Status, &job.Attempts, &job.MaxAttempts, &job.TenantID, &job.CreatedAt, &job.UpdatedAt); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if err := json.Unmarshal([]byte(payloadStr), &job.Payload); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE queue_jobs SET status = 'processing', attempts = attempts + 1, updated_at = NOW() WHERE id = $1`,
		job.ID); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	return tx, &job, nil
}

// MarkDone marks a job as completed.
func MarkDone(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE queue_jobs SET status = 'completed', processed_at = NOW(), updated_at = NOW() WHERE id = $1`,
		id)
	return err
}

// MarkFailed marks a job as failed or re-queues it for retry.
func MarkFailed(ctx context.Context, tx *sql.Tx, id uuid.UUID, attempts, maxAttempts int, errMsg string) error {
	status := "pending"
	if attempts >= maxAttempts {
		status = "failed"
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE queue_jobs SET status = $2, error_message = $3, scheduled_at = NOW() + (attempts * INTERVAL '5 minutes'), updated_at = NOW() WHERE id = $1`,
		id, status, errMsg)
	return err
}

// DeleteOldCompleted removes completed/failed jobs older than the retention window.
func DeleteOldCompleted(ctx context.Context, db *sql.DB, retention time.Duration) (int64, error) {
	result, err := db.ExecContext(ctx,
		`DELETE FROM queue_jobs WHERE status IN ('completed','failed') AND updated_at < NOW() - $1`,
		retention)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ParsePayload unmarshals the job payload into v.
func (j *Job) ParsePayload(v interface{}) error {
	return json.Unmarshal(j.Payload, v)
}

func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
