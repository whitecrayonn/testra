# Testra Troubleshooting Guide

## General Triage

1. Capture environment, timestamp, endpoint/job, request ID, tenant scope, and symptom.
2. Check service health and recent deploys.
3. Check structured logs, metrics, traces, queue state, and database health in that order.
4. Determine whether the issue is isolated, tenant-wide, regional, or global.
5. Mitigate safely, preserve evidence, and document the incident.

## Symptom Matrix

| Symptom | First checks | Likely areas |
|---|---|---|
| `401` | token expiry/signature, auth middleware, clock skew | identity/config |
| `403` | membership, role, scope, route permission | RBAC/tenancy |
| `404` for known resource | tenant scope and visibility rules | authorization/query |
| `409` on create | uniqueness and duplicate request | domain/database |
| API latency | dependency latency, DB pool, queue backlog | API/PostgreSQL/Redis |
| Worker backlog | Redis health, retries, dead letters, worker capacity | queue/worker |
| Missing analytics | ingestion status, ClickHouse writes, retention | results/ClickHouse |
| Email not received locally | Mailpit health and SMTP configuration | SMTP/config |
| Frontend API errors | base URL, CORS, token, network response | web/API |

## Safety Rules

Do not disable authentication, tenant checks, TLS, rate limits, or audit behavior as a first response. Do not print tokens, passwords, API keys, request bodies containing secrets, or customer source code into logs.

## Recovery Escalation

Escalate immediately for cross-tenant exposure, credential compromise, data loss, sustained production unavailability, migration failure, or unexplained authorization bypass. Preserve logs and timestamps before rotating or deleting anything.

## Closure

Record root cause, customer impact, detection gap, mitigation, permanent fix, and follow-up owner. Add a regression test or documentation update where appropriate.
