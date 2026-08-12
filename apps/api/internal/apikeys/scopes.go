package apikeys

import (
	"strings"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

// AllowedScopes is the registry of API-key scopes that may be requested at
// create time. Any scope outside this set is rejected so that keys cannot be
// created with arbitrary permission strings. The registry intentionally mirrors
// the permission namespace used by the RBAC system but is restricted to the
// actions that make sense for long-lived automation keys.
var AllowedScopes = map[string]bool{
	"runs:ingest":         true,
	"results:write":       true,
	"results:read":        true,
	"integrations:read":   true,
	"integrations:create": true,
	"analytics:read":      true,
	"intelligence:read":   true,
}

// ValidateScopes returns ErrInvalidInput if any requested scope is not in the
// allowed registry or if the list is empty.
func ValidateScopes(scopes []string) error {
	if len(scopes) == 0 {
		return sharederrors.ErrInvalidInput
	}
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s == "" || !AllowedScopes[s] {
			return sharederrors.ErrInvalidInput
		}
	}
	return nil
}
