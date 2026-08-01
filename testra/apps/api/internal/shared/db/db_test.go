package db

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestBeginTx_RollsBackWhenSetLocalTenantIDFails(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer mockDB.Close()

	db := Wrap(mockDB)

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ctx := WithTenantID(context.Background(), tenantID)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL app.tenant_id = '" + tenantID.String() + "'")).
		WillReturnError(errors.New("connection lost"))
	mock.ExpectRollback()

	tx, err := db.BeginTx(ctx, nil)
	if err == nil {
		t.Fatal("expected error from BeginTx when SetLocalTenantID fails")
	}
	if tx != nil {
		t.Fatalf("expected nil transaction, got %v", tx)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled sqlmock expectations: %v", err)
	}
}
