# Testra Monitoring and Logging Guide

## Observability Goals

Observe availability, latency, correctness, tenant isolation, queue health, storage health, and security events without collecting customer source code or sensitive payloads.

## Required Signals

- **Metrics:** request rate, error rate, latency percentiles, database pool saturation, queue depth/retries, worker throughput, ClickHouse ingestion failures, storage errors, authentication failures, and rate-limit events.
- **Logs:** structured JSON with timestamp, level, service, environment, request ID, operation, outcome, duration, and safe tenant/resource identifiers where permitted.
- **Traces:** HTTP-to-use-case-to-database/queue/external dependency spans with secret-free attributes.
- **Dashboards:** API health, dependency health, ingestion, authentication/security, and business availability.

The required observability stack is OpenTelemetry, Prometheus, Grafana, Loki, and structured Go logging. MVP runs these through AWS-managed or low-operations equivalents where practical; Beta and Enterprise add multi-AZ/centralized retention controls.

## Logging Rules

Never log passwords, JWTs, API keys, MFA secrets, reset tokens, raw request bodies, customer source code, or raw API collection payloads. Hash or redact identifiers when full identity is not required. Use stable error codes for aggregation.

## Alerting

Alert on sustained error-budget burn, elevated p95/p99 latency, database saturation, queue backlog/dead letters, failed migrations, backup failures, authentication attack patterns, and cross-tenant authorization signals. Alerts require severity, owner, runbook link, and escalation path.

## Retention

Retention is fixed by ADR-005: application logs are 30 days hot and 90 days archived; metrics are 15 months; traces are 14 days; audit records are at least 2 years for MVP/Beta and 7 years for Enterprise governance. Logs and telemetry must exclude secrets, source code, and raw API payloads.

## Operational Review

Every release checks dashboards and alerts. Alerting must include the ADR-008 API latency/query targets, queue and ClickHouse ingestion thresholds, backup age/failure, authentication abuse, and cross-tenant authorization signals. Every incident reviews whether detection was timely, actionable, and privacy-safe.
