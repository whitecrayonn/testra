package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type contextKey int

const (
	txKey contextKey = iota
	connKey
	tenantKey
	lookupUserKey
	apiKeyWorkspaceKey
)

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantKey, tenantID)
}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(tenantKey).(uuid.UUID)
	return v, ok
}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func TxFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txKey).(*sql.Tx)
	return tx
}

func WithConn(ctx context.Context, conn *sql.Conn) context.Context {
	return context.WithValue(ctx, connKey, conn)
}

func ConnFromContext(ctx context.Context) *sql.Conn {
	conn, _ := ctx.Value(connKey).(*sql.Conn)
	return conn
}

func WithLookupUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, lookupUserKey, userID)
}

func LookupUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(lookupUserKey).(uuid.UUID)
	return v, ok
}

// WithAPIKeyWorkspaceID records the workspace an authenticating API key is
// scoped to. Unlike TenantID (the key's organization), this lets callers
// enforce that a key only acts within the single workspace it was
// provisioned for, not any workspace in the organization.
func WithAPIKeyWorkspaceID(ctx context.Context, workspaceID uuid.UUID) context.Context {
	return context.WithValue(ctx, apiKeyWorkspaceKey, workspaceID)
}

func APIKeyWorkspaceIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(apiKeyWorkspaceKey).(uuid.UUID)
	return v, ok
}
