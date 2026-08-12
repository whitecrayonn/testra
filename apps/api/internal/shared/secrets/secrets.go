package secrets

import (
	"fmt"
	"os"
)

// Provider abstracts how the application retrieves sensitive values such as
// SMTP passwords, API keys, or signing keys. Implementations may read from
// environment variables, a file, an external secret manager, or a development
// in-memory store.
type Provider interface {
	// Get returns the value for the given secret key/name.
	Get(key string) (string, error)
}

// EnvProvider retrieves secret values from environment variables.
type EnvProvider struct{}

// NewEnvProvider returns a Provider that reads from environment variables.
func NewEnvProvider() Provider {
	return &EnvProvider{}
}

func (p *EnvProvider) Get(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("secret %q is not set", key)
	}
	return v, nil
}
