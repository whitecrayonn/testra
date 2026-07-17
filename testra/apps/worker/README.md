# Testra Worker

The **Testra Worker** is an optional standalone Go worker entrypoint. It is currently a placeholder; the main API monolith handles all background work in-process for the MVP.

## Planned role

- Asynchronous task processing (e.g., Asynq / Redis-backed queues).
- Bulk ingestion, report generation, analytics ETL, and notification delivery.

## Current state

- No worker process is defined.
- `apps/api` performs all work synchronously in HTTP handlers.

## When to use this directory

Create a standalone worker when the async pipeline is implemented (planned for Phase 5 / analytics).

## Canonical documentation

- [Engineering Handbook](../../docs/BIBLICAL_TESTRA.md)
- [Engineering Roadmap](../../docs/engineering/ROADMAP.md)
