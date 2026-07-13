package rbac

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type SQLPermissionLoader struct {
	db *sql.DB
}

func NewSQLPermissionLoader(db *sql.DB) *SQLPermissionLoader {
	return &SQLPermissionLoader{db: db}
}

func (l *SQLPermissionLoader) LoadPermissions(ctx context.Context, userID uuid.UUID, scopeType string, scopeID uuid.UUID) ([]string, error) {
	rows, err := l.db.QueryContext(ctx,
		`SELECT DISTINCT p.name
		 FROM role_assignments ra
		 JOIN role_permissions rp ON ra.role_id = rp.role_id
		 JOIN permissions p ON rp.permission_id = p.id
		 WHERE ra.user_id = $1 AND ra.scope_type = $2 AND ra.scope_id = $3`,
		userID, scopeType, scopeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}
