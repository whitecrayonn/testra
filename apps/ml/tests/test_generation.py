import os

from fastapi.testclient import TestClient

from api.main import app

os.environ.setdefault("ML_API_KEY", "test-key")

client = TestClient(app)

API_KEY_HEADERS = {"X-API-Key": "test-key"}

SAMPLE_CSV = (
    b"Title,Steps,Expected\n"
    b"Login succeeds with valid credentials,Enter valid email and password and submit,Dashboard loads\n"
    b"Login fails with wrong password,Enter valid email and wrong password and submit,Error message shown\n"
)


def _upload(content: bytes, filename: str = "cases.csv", context: str = ""):
    return client.post(
        "/generate-test-cases-from-file",
        files={"file": (filename, content, "text/csv")},
        data={"context": context},
        headers=API_KEY_HEADERS,
    )


def test_generate_without_gemini_key_returns_503(monkeypatch):
    monkeypatch.delenv("GEMINI_API_KEY", raising=False)
    res = _upload(SAMPLE_CSV)
    assert res.status_code == 503
    assert "not configured" in res.json()["detail"].lower()


def test_generate_rejects_unsupported_file_type():
    res = _upload(b"not a spreadsheet", filename="notes.txt")
    assert res.status_code == 400


def test_generate_success_maps_rows_to_cases(monkeypatch):
    def fake_call_gemini(prompt: str) -> dict:
        assert "Login succeeds" in prompt
        return {
            "cases": [
                {
                    "title": "Login succeeds with valid credentials",
                    "description": "Verify a user can log in with correct credentials.",
                    "preconditions": "User account exists",
                    "priority": "high",
                    "tags": ["auth"],
                    "steps": [
                        {
                            "action": "Enter valid email and password and submit",
                            "expected": "Dashboard loads",
                            "test_data": "user@example.com / correct-password",
                        }
                    ],
                    "row_index": 0,
                },
                {
                    "title": "Login fails with wrong password",
                    "description": "Verify login is rejected with an incorrect password.",
                    "preconditions": "User account exists",
                    "priority": "medium",
                    "tags": ["auth"],
                    "steps": [
                        {
                            "action": "Enter valid email and wrong password and submit",
                            "expected": "Error message shown",
                            "test_data": "",
                        }
                    ],
                    "row_index": 1,
                },
            ],
            "skipped": [],
        }

    monkeypatch.setattr("api.generation._call_gemini", fake_call_gemini)
    monkeypatch.setenv("GEMINI_API_KEY", "fake-key-for-test")

    res = _upload(SAMPLE_CSV, context="Login feature")
    assert res.status_code == 200
    body = res.json()
    assert body["row_count"] == 2
    assert len(body["cases"]) == 2
    assert body["cases"][0]["priority"] == "high"
    assert body["cases"][1]["steps"][0]["expected"] == "Error message shown"
    assert body["skipped_rows"] == []


def test_generate_reports_skipped_rows_instead_of_fabricating(monkeypatch):
    def fake_call_gemini(prompt: str) -> dict:
        return {
            "cases": [],
            "skipped": [{"row_index": 0, "reason": "Row is empty and has no test intent."}],
        }

    monkeypatch.setattr("api.generation._call_gemini", fake_call_gemini)
    monkeypatch.setenv("GEMINI_API_KEY", "fake-key-for-test")

    res = _upload(SAMPLE_CSV)
    assert res.status_code == 200
    body = res.json()
    assert body["cases"] == []
    assert body["skipped_rows"] == [{"row": 0, "reason": "Row is empty and has no test intent."}]


def test_generate_skips_case_that_fails_validation(monkeypatch):
    def fake_call_gemini(prompt: str) -> dict:
        return {
            "cases": [
                {
                    "title": "Case with no steps",
                    "description": "",
                    "preconditions": "",
                    "priority": "medium",
                    "tags": [],
                    "steps": [],  # invalid: at least 1 step required
                    "row_index": 0,
                }
            ],
            "skipped": [],
        }

    monkeypatch.setattr("api.generation._call_gemini", fake_call_gemini)
    monkeypatch.setenv("GEMINI_API_KEY", "fake-key-for-test")

    res = _upload(SAMPLE_CSV)
    assert res.status_code == 200
    body = res.json()
    assert body["cases"] == []
    assert len(body["skipped_rows"]) == 1
    assert body["skipped_rows"][0]["row"] == 0
