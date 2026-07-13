# Testra Release Checklist

## Planning

- [ ] Release scope maps to an approved phase and owning module.
- [ ] Compatibility impact and migration notes are documented.
- [ ] OpenAPI, database, security, and operational docs are updated.
- [ ] Release notes identify user-visible changes, risks, and rollback strategy.

## Validation

- [ ] Unit, integration, contract, and applicable end-to-end tests pass.
- [ ] Lint, typecheck, build, and security scans pass.
- [ ] Migration up/down behavior and production duration are reviewed.
- [ ] Tenant isolation and authorization tests pass.
- [ ] Observability dashboards and alerts cover the change.
- [ ] Backup/restore implications are addressed.

## Staging

- [ ] Immutable artifact is deployed to staging.
- [ ] Migrations run through the deployment pipeline.
- [ ] Smoke tests pass for authentication, tenancy, core API, queues, and relevant UI.
- [ ] No unexplained errors, latency regression, or backlog growth during soak.

## Production

- [ ] Production readiness checklist is approved.
- [ ] Promotion approval and incident coverage are confirmed.
- [ ] Same artifact digest is promoted.
- [ ] Post-deploy health checks and critical user journey checks pass.
- [ ] Release outcome, metrics, and any follow-up work are recorded.

## Rollback / Forward Fix

- [ ] Rollback trigger is defined.
- [ ] Schema compatibility is confirmed.
- [ ] Forward-fix procedure is defined for irreversible migrations.
- [ ] Customer and internal communication path is ready.
