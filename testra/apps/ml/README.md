# Testra ML

The **Testra ML** service is a Python 3.12+ FastAPI service for tenant-scoped machine learning inference. It is currently a skeleton with only a `/health` endpoint.

## Current state

- `apps/ml/api/main.py` exposes a `/health` endpoint.
- No model training or inference endpoints are wired yet.
- The service is not integrated into the main deployment pipeline.

## Planned capabilities (Phase 6)

- Flaky test detection using time-series variance scoring.
- Failure classification (rule-based filtering + DBSCAN/HDBSCAN clustering, optional XGBoost).
- Risk/health scores with logistic regression / XGBoost and SHAP explainability.
- Release readiness thresholds and trend analysis.

## Principles

- No external LLM APIs.
- Models trained per tenant on that tenant's data only.
- Inputs limited to test metadata and results; never source code or secrets.

## Running locally

See `docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md` for Python environment setup.

## Canonical documentation

- [Engineering Handbook](../../docs/BIBLICAL_TESTRA.md)
- [Implementation Phases](../../docs/engineering/PHASES.md)
- [Local Development Guide](../../docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md)
