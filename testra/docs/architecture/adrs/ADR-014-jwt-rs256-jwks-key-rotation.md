# ADR-014: JWT signing with RS256, JWKS endpoint, and key rotation

## Status

Accepted

## Context

The API currently signs access tokens with `golang-jwt` `HS256` and a single `JWT_SECRET` symmetric key. This creates several production risks:

1. **Key rotation is hard.** Any service that verifies tokens needs the full signing secret, so compromise of any verifier leaks the signing key.
2. **No `aud`/`iss` validation.** Tokens are not bound to an issuer or audience, increasing replay and token substitution risk.
3. **No key identifier (`kid`).** All tokens use the same symmetric key with no header identifying which key was used.
4. **No JWKS endpoint.** External callers and internal services have no standard way to obtain public verification keys.

The refresh-token subsystem already uses token families and reuse detection; modernizing access-token signing is the remaining prerequisite for a production-grade authentication architecture.

## Decision

1. Adopt **RS256** asymmetric signing for access tokens.
2. Introduce a `TokenManager` component that encapsulates signing, verification, JWKS publication, and key-set rotation.
3. Each token header carries a `kid` identifying the RSA key pair used to sign it.
4. Verification uses the public key that matches the token's `kid` from a configurable key set.
5. Tokens include `iss` and `aud` registered claims and are validated against configured values.
6. Expose `GET /.well-known/jwks.json` returning the public keys in JWK format.
7. Production configuration supplies:
   - `JWT_ISSUER` (e.g. `https://api.testra.io`)
   - `JWT_AUDIENCE` (e.g. `testra-api`)
   - `JWT_PRIVATE_KEY` (PEM-encoded RSA private key, one-time secret)
   - `JWT_PUBLIC_KEYS` (comma-separated PEM-encoded public keys, including the current public key and any previously rotated keys for verification)
8. The `TokenManager` keeps a verification key set in a thread-safe map. Rotation is performed by replacing the signing key and appending the previous public key to the verification set; no previously issued valid token is rejected until its natural expiry, and no restart is required.
9. `middleware.Auth` receives the `TokenManager` instead of the raw `JWTSecret`.
10. `identity.Service` receives the `TokenManager` and uses it to sign access tokens.

## Consequences

- Access-token signing and verification are now cryptographically separated: verifiers never need the private key.
- Key rotation can be done without invalidating existing sessions, as long as the previous public key remains in `JWT_PUBLIC_KEYS`.
- External and internal clients can retrieve public keys via the JWKS endpoint.
- `aud` and `iss` validation reduce token misuse across environments.
- `kid` support enables smooth revocation of compromised signing keys by removing their public key from `JWT_PUBLIC_KEYS`.
- The `JWT_SECRET` environment variable is removed; operators must provision RSA key material. This is a breaking configuration change.
- Unit and integration tests use an in-memory RSA key pair generated per test run.

## Validation

- `go test ./internal/shared/jwt/...` covers signing, verification, `aud`/`iss`/`kid`/`exp` failure modes, and JWKS output.
- `go test ./internal/identity/...` ensures login/refresh/Me still issue and accept access tokens.
- `go test ./internal/shared/middleware/...` verifies the auth middleware rejects invalid tokens and accepts valid RS256 tokens.
