# ADR-016: SMTP authentication and secret provider abstraction

## Status

Accepted

## Context

The notification and identity services sent email through `smtp.SendMail` with a `nil` `Auth` argument. `SMTPConfig` also carried a `Password` field directly in memory, which means credentials were either absent or hard to rotate, and could not be sourced from a secret manager in production.

## Decision

1. Introduce a minimal `secrets.Provider` interface in `apps/api/internal/shared/secrets`.
   - The first implementation is `EnvProvider`, which reads values from environment variables.
   - The interface can be replaced with a Vault, environment files or a local secrets store, or single-Ubuntu-VPS systemd services secrets backend without touching service code.
2. Extend `SMTPConfig` with:
   - `Username` — the SMTP account name.
   - `SecretProvider` — a `secrets.Provider` used to resolve the password.
   - `PasswordSecret` — the key/name passed to the provider (default `SMTP_PASSWORD`).
   - `Password` kept as a deprecated fallback for local development and existing tests.
3. At dispatch time, resolve the password from the secret provider when both `Username` and `PasswordSecret` are configured. Build `smtp.PlainAuth` when both username and resolved password are non-empty.
4. Add new environment variables:
   - `SMTP_USERNAME`
   - `SMTP_PASSWORD_SECRET` (defaults to `SMTP_PASSWORD`)
5. Wire the secret provider into `server.Config` so the worker process and API server share the same abstraction.

## Consequences

- SMTP credentials can now be rotated externally (e.g., via a secret manager or single-Ubuntu-VPS systemd services secret mounted as an env var) without code changes.
- `smtp.SendMail` is invoked with an authenticated `PlainAuth` when configured, preventing open-relay misconfigurations.
- The API, worker, and identity services share a single `SecretProvider` abstraction for future credential management.
- Passwords are no longer required to live in the config struct when `SecretProvider` is used.

## Validation

- `go build ./...` and `go vet ./...` pass in `apps/api`.
- `go test ./internal/notification/...` includes `TestDispatchEmailUsesPlainAuthFromSecretProvider`.
- `cmd/worker/main.go` and `cmd/api/main.go` both build with the updated `notification.NewModule` and `SMTPConfig` signatures.
