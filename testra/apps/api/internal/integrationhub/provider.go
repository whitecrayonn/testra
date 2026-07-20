package integrationhub

import (
	"context"
	"fmt"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

// Provider is the common interface for every Integration Hub integration.
type Provider interface {
	// Type returns the provider type this implementation handles.
	Type() IntegrationType
	// Validate checks that the config map contains the required fields without
	// making any external calls.
	Validate(cfg map[string]string) error
	// Test performs a lightweight external connection test and returns an
	// external reference if one is created by the call.
	Test(ctx context.Context, i *Integration) (string, error)
	// Health reports the provider-specific health status: "healthy", "degraded" or "error".
	Health(ctx context.Context, i *Integration) (string, error)
	// Send dispatches the payload to the external system.
	Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error)
}

var providers = map[IntegrationType]Provider{}

func init() {
	RegisterProvider(&jiraProvider{})
	RegisterProvider(&githubProvider{})
	RegisterProvider(&gitlabProvider{})
	RegisterProvider(&bitbucketProvider{})
	RegisterProvider(&azureDevOpsProvider{})
	RegisterProvider(&linearProvider{})
	RegisterProvider(&slackProvider{})
	RegisterProvider(&discordProvider{})
	RegisterProvider(&smtpProvider{})
	RegisterProvider(&webhookProvider{})
}

// RegisterProvider registers a Provider implementation. Panics on duplicate types.
func RegisterProvider(p Provider) {
	providers[p.Type()] = p
}

// ProviderFor returns the registered Provider for a type or ErrInvalidInput.
func ProviderFor(t IntegrationType) (Provider, error) {
	p, ok := providers[t]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported integration type %s", sharederrors.ErrInvalidInput, t)
	}
	return p, nil
}

// dispatch routes a payload to the correct provider.
func dispatch(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	p, err := ProviderFor(i.Type)
	if err != nil {
		return "", err
	}
	return p.Send(ctx, i, payload)
}

// testIntegration performs a connection test through the correct provider.
func testIntegration(ctx context.Context, i *Integration) (string, error) {
	p, err := ProviderFor(i.Type)
	if err != nil {
		return "", err
	}
	return p.Test(ctx, i)
}

// healthIntegration returns the health status from the correct provider.
func healthIntegration(ctx context.Context, i *Integration) (string, error) {
	p, err := ProviderFor(i.Type)
	if err != nil {
		return "error", err
	}
	return p.Health(ctx, i)
}

// validateProviderConfig validates the config for a provider type.
func validateProviderConfig(t IntegrationType, cfg map[string]string) error {
	p, err := ProviderFor(t)
	if err != nil {
		return err
	}
	return p.Validate(cfg)
}
