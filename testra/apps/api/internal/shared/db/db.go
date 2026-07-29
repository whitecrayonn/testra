package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	return db, nil
}

// DBTX is the subset of *sql.DB/*sql.Tx/*sql.Conn methods that repositories use.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// BeginTxer is implemented by *sql.DB, *sql.Conn and *DB.
type BeginTxer interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// DB wraps a *sql.DB and transparently uses a per-request *sql.Conn or *sql.Tx
// when one is present in the context. This lets middleware set the connection
// local tenant id (app.tenant_id) once per request and have all repository
// calls on that request use the same connection.
type DB struct {
	db *sql.DB
}

// SetSessionTenantID sets app.tenant_id for the current database session.
// PostgreSQL SET does not support parameters, so the safe UUID is interpolated.
func SetSessionTenantID(ctx context.Context, exec DBTX, tenantID uuid.UUID) error {
	_, err := exec.ExecContext(ctx, fmt.Sprintf("SET app.tenant_id = '%s'", tenantID.String()))
	return err
}

// SetLocalTenantID sets app.tenant_id for the current transaction.
// PostgreSQL SET LOCAL does not support parameters, so the safe UUID is interpolated.
func SetLocalTenantID(ctx context.Context, exec DBTX, tenantID uuid.UUID) error {
	_, err := exec.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID.String()))
	return err
}

// SetLookupKeyHash sets app.lookup_key_hash for API-key lookup policies.
func SetLookupKeyHash(ctx context.Context, exec DBTX, hash string) error {
	_, err := exec.ExecContext(ctx, fmt.Sprintf("SET app.lookup_key_hash = '%s'", hash))
	return err
}

func Wrap(db *sql.DB) *DB {
	return &DB{db: db}
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	if conn := ConnFromContext(ctx); conn != nil {
		return conn.ExecContext(ctx, query, args...)
	}
	return d.db.ExecContext(ctx, query, args...)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	if conn := ConnFromContext(ctx); conn != nil {
		return conn.QueryContext(ctx, query, args...)
	}
	return d.db.QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx := TxFromContext(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	if conn := ConnFromContext(ctx); conn != nil {
		return conn.QueryRowContext(ctx, query, args...)
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	var tx *sql.Tx
	var err error

	if txInCtx := TxFromContext(ctx); txInCtx != nil {
		// Already inside a transaction; reuse it. This avoids trying to start
		// a nested transaction, which database/sql does not support.
		return txInCtx, nil
	}

	if conn := ConnFromContext(ctx); conn != nil {
		tx, err = conn.BeginTx(ctx, opts)
	} else {
		tx, err = d.db.BeginTx(ctx, opts)
	}
	if err != nil {
		return nil, err
	}

	if tenantID, ok := TenantIDFromContext(ctx); ok {
		_ = SetLocalTenantID(ctx, tx, tenantID)
	}
	return tx, nil
}

// Ensure DB implements the interfaces it is used as.
var _ DBTX = (*DB)(nil)
var _ BeginTxer = (*DB)(nil)
