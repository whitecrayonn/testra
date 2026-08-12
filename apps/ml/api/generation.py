"""LLM-backed test case generation from an uploaded spreadsheet.

This is the one intentional exception to Testra's "no external LLM" principle
(see docs/BIBLICAL_TESTRA.md). It is opt-in: the endpoint refuses to run
unless GEMINI_API_KEY is configured, and every generated case is written
upstream as pending_review, never auto-activated.
"""

import io
import json
import os
from typing import Any

import pandas as pd
from fastapi import APIRouter, Depends, File, Form, HTTPException, UploadFile, status
from pydantic import BaseModel, Field, ValidationError

from .auth import verify_api_key

router = APIRouter()

MAX_ROWS = 200
MAX_FILE_BYTES = 5 * 1024 * 1024  # 5 MB
GEMINI_MODEL = "gemini-flash-latest"
PRIORITIES = {"low", "medium", "high", "critical"}


class GeneratedStep(BaseModel):
    action: str = Field(..., min_length=1, max_length=2000)
    expected: str = Field(..., min_length=1, max_length=2000)
    test_data: str = Field(default="", max_length=2000)


class GeneratedCase(BaseModel):
    title: str = Field(..., min_length=1, max_length=300)
    description: str = Field(default="", max_length=2000)
    preconditions: str = Field(default="", max_length=2000)
    priority: str = Field(default="medium")
    tags: list[str] = Field(default_factory=list, max_length=20)
    steps: list[GeneratedStep] = Field(..., min_length=1, max_length=30)

    def normalized_priority(self) -> str:
        p = self.priority.lower().strip()
        return p if p in PRIORITIES else "medium"


class SkippedRow(BaseModel):
    row: int
    reason: str


class GenerateFromFileResponse(BaseModel):
    cases: list[GeneratedCase]
    skipped_rows: list[SkippedRow]
    row_count: int


def _read_rows(filename: str, content: bytes) -> pd.DataFrame:
    lower = filename.lower()
    try:
        if lower.endswith(".csv"):
            return pd.read_csv(io.BytesIO(content))
        if lower.endswith(".xlsx") or lower.endswith(".xls"):
            return pd.read_excel(io.BytesIO(content), engine="openpyxl")
    except Exception as exc:  # noqa: BLE001 - surfaced to the caller as a 400
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Could not parse file: {exc}",
        ) from exc
    raise HTTPException(
        status_code=status.HTTP_400_BAD_REQUEST,
        detail="Unsupported file type; upload a .csv or .xlsx file.",
    )


def _rows_to_records(df: pd.DataFrame) -> list[dict[str, Any]]:
    df = df.dropna(how="all")
    records = df.head(MAX_ROWS).fillna("").to_dict(orient="records")
    return [{str(k): str(v) for k, v in record.items()} for record in records]


SCHEMA_INSTRUCTIONS = """You convert spreadsheet rows describing software behavior into structured QA test cases.

For EACH input row, output one test case object with this exact shape:
{
  "title": string (required, short imperative name),
  "description": string (1-2 sentence summary, may be empty),
  "preconditions": string (setup required before the test, may be empty),
  "priority": one of "low", "medium", "high", "critical" (infer from context, default "medium"),
  "tags": array of short strings (e.g. feature/module name, may be empty),
  "steps": array of at least 1 object {"action": string, "expected": string, "test_data": string},
  "row_index": integer, the 0-based index of the input row this case was built from
}

Rules:
- Do not invent behavior that is not implied by the row's own content.
- If a row is empty, unintelligible, or has no discernible test intent, OMIT it from "cases" and instead add
  {"row_index": <index>, "reason": <short reason>} to a separate "skipped" array.
- Respond with strict JSON only: {"cases": [...], "skipped": [...]}. No prose, no markdown fences.
"""


def _build_prompt(records: list[dict[str, Any]], context: str) -> str:
    parts = [SCHEMA_INSTRUCTIONS]
    if context:
        parts.append(f"\nContext supplied by the user about this batch: {context}\n")
    parts.append("\nInput rows (JSON array, one object per spreadsheet row):")
    parts.append(json.dumps(records, ensure_ascii=False))
    return "\n".join(parts)


def _call_gemini(prompt: str) -> dict[str, Any]:
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="LLM not configured: GEMINI_API_KEY is not set on the ML service.",
        )

    from google import genai
    from google.genai import types

    client = genai.Client(api_key=api_key)

    try:
        response = client.models.generate_content(
            model=GEMINI_MODEL,
            contents=prompt,
            config=types.GenerateContentConfig(temperature=0.2, response_mime_type="application/json"),
        )
    except Exception as exc:  # noqa: BLE001 - classified into a stable status code below
        message = str(exc)
        if "429" in message or "quota" in message.lower() or "rate" in message.lower():
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail="AI generation is rate-limited right now. Try again shortly.",
            ) from exc
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail=f"LLM call failed: {message}",
        ) from exc

    try:
        return json.loads(response.text)
    except (ValueError, AttributeError) as exc:
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail="LLM did not return valid JSON.",
        ) from exc


@router.post("/generate-test-cases-from-file", response_model=GenerateFromFileResponse)
def generate_test_cases_from_file(
    file: UploadFile = File(...),
    context: str = Form(default=""),
    _=Depends(verify_api_key),
) -> GenerateFromFileResponse:
    content = file.file.read(MAX_FILE_BYTES + 1)
    if len(content) > MAX_FILE_BYTES:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="File too large; max 5 MB.")

    df = _read_rows(file.filename or "upload.csv", content)
    records = _rows_to_records(df)
    if not records:
        return GenerateFromFileResponse(cases=[], skipped_rows=[], row_count=0)

    raw = _call_gemini(_build_prompt(records, context))

    cases: list[GeneratedCase] = []
    skipped: list[SkippedRow] = []

    for raw_case in raw.get("cases", []):
        row_index = raw_case.get("row_index", -1) if isinstance(raw_case, dict) else -1
        try:
            case = GeneratedCase.model_validate(raw_case)
            case.priority = case.normalized_priority()
            cases.append(case)
        except ValidationError as exc:
            skipped.append(
                SkippedRow(row=row_index, reason=f"Model output failed validation: {exc.errors()[0]['msg']}")
            )

    for raw_skip in raw.get("skipped", []):
        if isinstance(raw_skip, dict):
            skipped.append(
                SkippedRow(
                    row=int(raw_skip.get("row_index", -1)),
                    reason=str(raw_skip.get("reason", "Not enough information to build a test case.")),
                )
            )

    return GenerateFromFileResponse(cases=cases, skipped_rows=skipped, row_count=len(records))
