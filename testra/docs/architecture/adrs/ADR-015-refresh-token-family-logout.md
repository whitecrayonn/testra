# ADR-015: Refresh token families, reuse detection, and logout

## Status

Accepted

## Context

Access-token signing has been modernized (ADR-014), but the refresh-token layer still needs production-grade security. The existing `refresh_tokens` table already stores `family_id`, `revoked_at`, and `replaced_by`, but the service layer does not fully exploit those columns for:

1. **Token family binding**: every login/registration should create a new token family.
2. **Reuse detection**: presenting a refresh token that has already been consumed or revoked must invalidate the entire token family (descendants included).
3. **Secure logout**: users need endpoints to log out a single device (revoke its refresh-token family) and to log out all devices (revoke every refresh token for the user).

## Decision

1. Each successful `Register` or `Login` creates a new `RefreshToken` with a fresh `family_id`.
2. Each successful `RefreshToken` call:
   - validates the stored token is not revoked or expired,
   - issues a new refresh token within the **same** `family_id`,
   - revokes the old token and sets `replaced_by` to the new token's ID,
   - returns a new access token and the new refresh token.
3. If a refresh token is presented that is already revoked or replaced, the system treats it as a reuse attempt and revokes **all tokens in the family**, returning `ErrTokenRevoked`.
4. `POST /auth/logout` accepts a `refresh_token`, looks it up, and revokes its family. The endpoint is intentionally idempotent (a second call succeeds with `logged_out`).
5. `POST /auth/logout-all` requires authentication and revokes every refresh token for the authenticated user.
6. Logout operates only on refresh tokens; access tokens remain short-lived and are not maintained in a revocation list.

## Consequences

- Stolen refresh tokens cannot be replayed indefinitely; reuse of an old token burns the whole family.
- Users can terminate sessions per-device or globally.
- The refresh-token table remains the source of truth; no additional token blacklist is required.
- A race condition where two clients simultaneously refresh is mitigated by the same `family_id` chain: one wins and revokes the old token; the other receives `ErrTokenRevoked`.
- The access token remains valid until expiry after logout, consistent with a stateless short-lived JWT design.

## Validation

- `go test ./internal/identity/...` covers:
  - successful refresh within a family,
  - reuse detection and family revocation,
  - single-device logout,
  - logout-all across multiple families.
- `go test ./...` and `go vet ./...` pass in `apps/api`.
