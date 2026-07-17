from fastapi import FastAPI
from pydantic import BaseModel, Field

app = FastAPI(title="Testra ML Service", version="0.0.0")


class RunHistoryPoint(BaseModel):
    status: str = Field(..., pattern="^(passed|failed|skipped|blocked|timeout)$")
    duration_ms: int = Field(..., ge=0)
    date: str


class PredictFlakyRequest(BaseModel):
    test_case_id: str | None = None
    test_case_title: str = ""
    history: list[RunHistoryPoint] = []


class PredictFlakyResponse(BaseModel):
    flakiness_score: float = Field(..., ge=0.0, le=1.0)
    confidence: float = Field(..., ge=0.0, le=1.0)
    explanation: str


class ClassifyFailureRequest(BaseModel):
    error_message: str
    stack_trace: str = ""


class ClassifyFailureResponse(BaseModel):
    label: str
    confidence: float = Field(..., ge=0.0, le=1.0)
    explanation: str


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/predict-flaky", response_model=PredictFlakyResponse)
def predict_flaky(req: PredictFlakyRequest) -> PredictFlakyResponse:
    """Predict flakiness from recent run history without external LLMs."""
    history = req.history
    n = len(history)
    if n == 0:
        return PredictFlakyResponse(
            flakiness_score=0.0,
            confidence=0.5,
            explanation="No run history available; defaulting to non-flaky.",
        )

    normalized = [_normalize_status(h.status) for h in history]
    failures = sum(1 for s in normalized if s == "failed")
    passes = sum(1 for s in normalized if s == "passed")
    transitions = sum(1 for i in range(1, n) if normalized[i] != normalized[i - 1])

    fail_ratio = failures / n
    transition_ratio = transitions / max(n - 1, 1)
    flakiness = min(1.0, (fail_ratio + transition_ratio) / 2.0)

    total_duration = sum(h.duration_ms for h in history)
    avg_duration = total_duration / n

    confidence = min(1.0, 0.5 + n * 0.05)
    explanation = (
        f"{n} runs, {transitions} transitions, {failures} failures, "
        f"{passes} passes, avg duration {avg_duration:.0f}ms"
    )

    return PredictFlakyResponse(
        flakiness_score=round(flakiness, 4),
        confidence=round(confidence, 4),
        explanation=explanation,
    )


@app.post("/classify-failure", response_model=ClassifyFailureResponse)
def classify_failure(req: ClassifyFailureRequest) -> ClassifyFailureResponse:
    """Classify a failure message using rule-based keyword matching."""
    text = f"{req.error_message} {req.stack_trace}".lower()
    label, explanation = _keyword_classify(text)
    return ClassifyFailureResponse(label=label, confidence=0.75, explanation=explanation)


def _normalize_status(status: str) -> str:
    s = status.lower().strip()
    if s in {"passed", "pass", "success"}:
        return "passed"
    if s in {"failed", "fail", "failure"}:
        return "failed"
    if s in {"skipped", "skip", "pending"}:
        return "skipped"
    return "blocked"


def _keyword_classify(text: str) -> tuple[str, str]:
    rules = [
        ("timeout", ["timeout", "timed out", "deadline exceeded", "waitfor"], "Timeout or wait condition exceeded"),
        ("network", ["network", "connection", "econnrefused", "socket", "curl", "request failed"], "Network or connection error"),
        ("assertion", ["assert", "assertion", "expected", "actual", "equal", "to be"], "Assertion or expectation mismatch"),
        ("ui_element", ["selector", "element", "dom", "not found", "no such element", "locator"], "Missing UI element or selector"),
        ("authorization", ["permission", "unauthorized", "forbidden", "access denied", "401", "403"], "Authorization or permission problem"),
        ("database", ["sql", "database", "query", "constraint", "foreign key", "unique constraint"], "Database or SQL error"),
    ]
    for label, keywords, explanation in rules:
        if any(kw in text for kw in keywords):
            return label, explanation
    return "unknown", "Failure does not match a known keyword class"

