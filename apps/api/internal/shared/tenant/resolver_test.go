package tenant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestResolveOrgFromWorkspace(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer dbMock.Close()

	resolver := NewResolver(dbMock)
	wsID := uuid.MustParse("a0a0a0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0")
	orgID := uuid.MustParse("b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1")

	mock.ExpectQuery("SELECT organization_id FROM workspaces WHERE id = \\$1").
		WithArgs(wsID).
		WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))

	got, err := resolver.ResolveOrgFromWorkspace(context.Background(), wsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != orgID {
		t.Fatalf("expected %s, got %s", orgID, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveOrgFromWorkspaceNotFound(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer dbMock.Close()

	resolver := NewResolver(dbMock)
	wsID := uuid.MustParse("a0a0a0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0")

	mock.ExpectQuery("SELECT organization_id FROM workspaces WHERE id = \\$1").
		WithArgs(wsID).
		WillReturnError(sql.ErrNoRows)

	_, err = resolver.ResolveOrgFromWorkspace(context.Background(), wsID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCheckMembership(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer dbMock.Close()

	resolver := NewResolver(dbMock)
	userID := uuid.MustParse("c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2")
	orgID := uuid.MustParse("b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1")

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(orgID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := resolver.CheckMembership(context.Background(), userID, orgID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckMembershipDeny(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer dbMock.Close()

	resolver := NewResolver(dbMock)
	userID := uuid.MustParse("c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2")
	orgID := uuid.MustParse("b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1")

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(orgID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	if err := resolver.CheckMembership(context.Background(), userID, orgID); err == nil {
		t.Fatalf("expected membership denial")
	}
}
