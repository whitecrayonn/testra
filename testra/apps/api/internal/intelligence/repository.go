package intelligence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type SQLRepository struct {
	db DBTX
}

func NewSQLRepository(sqlDB *sql.DB) *SQLRepository {
	return &SQLRepository{db: db.Wrap(sqlDB)}
}

func (r *SQLRepository) CreatePrediction(ctx context.Context, p *FlakyPrediction) error {
	featuresJSON, _ := json.Marshal(p.Features)
	tcID := uuid.NullUUID{}
	if p.TestCaseID != nil {
		tcID = uuid.NullUUID{UUID: *p.TestCaseID, Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO flaky_predictions (id, workspace_id, test_case_id, test_case_title, flakiness_score, confidence, features, predicted_at, last_seen_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		p.ID, p.WorkspaceID, tcID, p.TestCaseTitle, p.FlakinessScore, p.Confidence, featuresJSON, p.PredictedAt, p.LastSeenAt, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *SQLRepository) ListPredictions(ctx context.Context, workspaceID uuid.UUID, minScore float64, limit int) ([]FlakyPrediction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, test_case_id, test_case_title, flakiness_score, confidence, features::text, predicted_at, last_seen_at, created_at, updated_at
		 FROM flaky_predictions
		 WHERE workspace_id = $1 AND flakiness_score >= $2
		 ORDER BY flakiness_score DESC, created_at DESC
		 LIMIT $3`,
		workspaceID, minScore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPredictions(rows)
}

func (r *SQLRepository) GetPrediction(ctx context.Context, id uuid.UUID) (*FlakyPrediction, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, test_case_id, test_case_title, flakiness_score, confidence, features::text, predicted_at, last_seen_at, created_at, updated_at
		 FROM flaky_predictions WHERE id = $1`, id)
	return scanPrediction(row)
}

func (r *SQLRepository) CreateCluster(ctx context.Context, c *FailureCluster) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO failure_clusters (id, workspace_id, cluster_label, pattern, sample_error, count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.WorkspaceID, c.ClusterLabel, c.Pattern, c.SampleError, c.Count, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *SQLRepository) ListClusters(ctx context.Context, workspaceID uuid.UUID, limit int) ([]FailureCluster, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, workspace_id, cluster_label, pattern, sample_error, count, created_at, updated_at
		 FROM failure_clusters
		 WHERE workspace_id = $1
		 ORDER BY count DESC, created_at DESC
		 LIMIT $2`,
		workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClusters(rows)
}

func (r *SQLRepository) IncrementCluster(ctx context.Context, workspaceID uuid.UUID, label string, sampleError string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE failure_clusters SET count = count + 1, sample_error = $4, updated_at = NOW()
		 WHERE workspace_id = $1 AND cluster_label = $2`,
		workspaceID, label, sampleError) // positional args: $4 is sample_error
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c := &FailureCluster{
			ID:           uuid.New(),
			WorkspaceID:  workspaceID,
			ClusterLabel: label,
			Pattern:      label,
			SampleError:  sampleError,
			Count:        1,
			CreatedAt:    nowUTC(),
			UpdatedAt:    nowUTC(),
		}
		return r.CreateCluster(ctx, c)
	}
	return nil
}

func (r *SQLRepository) GetTestCaseHistory(ctx context.Context, testCaseID uuid.UUID, limit int) ([]RunHistoryPoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, duration_ms, DATE(created_at)
		 FROM test_run_items
		 WHERE test_case_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		testCaseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []RunHistoryPoint
	for rows.Next() {
		var h RunHistoryPoint
		var date sql.NullTime
		if err := rows.Scan(&h.Status, &h.DurationMs, &date); err != nil {
			return nil, err
		}
		if date.Valid {
			h.Date = date.Time.Format("2006-01-02")
		}
		history = append(history, h)
	}
	return history, rows.Err()
}

func (r *SQLRepository) RunInTx(ctx context.Context, fn func(Repository) error) error {
	beginner, ok := r.db.(db.BeginTxer)
	if !ok {
		return fmt.Errorf("database handle does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if tenantID, ok := db.TenantIDFromContext(ctx); ok {
		_, _ = tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String())
	}
	txRepo := &SQLRepository{db: tx}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanPrediction(row rowScanner) (*FlakyPrediction, error) {
	var p FlakyPrediction
	var tcID sql.NullString
	var featuresStr string
	if err := row.Scan(&p.ID, &p.WorkspaceID, &tcID, &p.TestCaseTitle, &p.FlakinessScore, &p.Confidence, &featuresStr, &p.PredictedAt, &p.LastSeenAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	if tcID.Valid {
		id, _ := uuid.Parse(tcID.String)
		p.TestCaseID = &id
	}
	if featuresStr != "" {
		_ = json.Unmarshal([]byte(featuresStr), &p.Features)
	}
	if p.Features == nil {
		p.Features = make(map[string]interface{})
	}
	return &p, nil
}

func scanPredictions(rows *sql.Rows) ([]FlakyPrediction, error) {
	var list []FlakyPrediction
	for rows.Next() {
		p, err := scanPrediction(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *p)
	}
	return list, rows.Err()
}

func scanCluster(row rowScanner) (*FailureCluster, error) {
	var c FailureCluster
	if err := row.Scan(&c.ID, &c.WorkspaceID, &c.ClusterLabel, &c.Pattern, &c.SampleError, &c.Count, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, sharederrors.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func scanClusters(rows *sql.Rows) ([]FailureCluster, error) {
	var list []FailureCluster
	for rows.Next() {
		c, err := scanCluster(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *c)
	}
	return list, rows.Err()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
