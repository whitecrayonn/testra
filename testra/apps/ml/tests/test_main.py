import os

from fastapi.testclient import TestClient

from api.main import app

os.environ.setdefault("ML_API_KEY", "test-key")

client = TestClient(app)

API_KEY_HEADERS = {"X-API-Key": "test-key"}


def test_health():
    res = client.get("/health")
    assert res.status_code == 200
    assert res.json() == {"status": "ok"}


def test_predict_flaky_empty_history():
    res = client.post("/predict-flaky", json={"history": []}, headers=API_KEY_HEADERS)
    assert res.status_code == 200
    body = res.json()
    assert body["flakiness_score"] == 0.0
    assert 0.0 <= body["confidence"] <= 1.0


def test_predict_flaky_alternating_history_is_flaky():
    history = [
        {"status": "passed", "duration_ms": 100, "date": "2026-01-01"},
        {"status": "failed", "duration_ms": 120, "date": "2026-01-02"},
        {"status": "passed", "duration_ms": 110, "date": "2026-01-03"},
        {"status": "failed", "duration_ms": 130, "date": "2026-01-04"},
    ]
    res = client.post("/predict-flaky", json={"history": history}, headers=API_KEY_HEADERS)
    assert res.status_code == 200
    body = res.json()
    assert body["flakiness_score"] > 0.5
    assert 0.0 <= body["flakiness_score"] <= 1.0


def test_predict_flaky_stable_history_is_not_flaky():
    history = [
        {"status": "passed", "duration_ms": 100, "date": "2026-01-01"},
        {"status": "passed", "duration_ms": 105, "date": "2026-01-02"},
        {"status": "passed", "duration_ms": 102, "date": "2026-01-03"},
    ]
    res = client.post("/predict-flaky", json={"history": history}, headers=API_KEY_HEADERS)
    assert res.status_code == 200
    assert res.json()["flakiness_score"] == 0.0


def test_classify_failure_timeout():
    res = client.post(
        "/classify-failure",
        json={"error_message": "operation timed out after 30s", "stack_trace": ""},
        headers=API_KEY_HEADERS,
    )
    assert res.status_code == 200
    assert res.json()["label"] == "timeout"


def test_classify_failure_unknown():
    res = client.post(
        "/classify-failure",
        json={"error_message": "lorem ipsum dolor sit amet", "stack_trace": ""},
        headers=API_KEY_HEADERS,
    )
    assert res.status_code == 200
    assert res.json()["label"] == "unknown"
