from fastapi import FastAPI

app = FastAPI(title="Testra ML Service", version="0.0.0")


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}
