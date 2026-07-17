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
