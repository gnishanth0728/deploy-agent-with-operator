
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from openai import OpenAI

app = FastAPI(
    title="Kubernetes AI Agent",
    version="1.0.0",
    description="AI agent managed by Kubernetes Operator"
)

MODEL = os.getenv("OPENAI_MODEL", "gpt-4.1-mini")

SYSTEM_PROMPT = os.getenv(
    "AGENT_SYSTEM_PROMPT",
    "You are a helpful DevOps AI assistant. "
    "Help users understand Kubernetes, Docker, "
    "Linux and cloud infrastructure."
)

ROLE = os.getenv("AGENT_ROLE", "devops-assistant")


class ChatRequest(BaseModel):
    message: str


class ChatResponse(BaseModel):
    agent: str
    model: str
    response: str


@app.get("/")
async def root():
    return {
        "name": "Kubernetes AI Agent",
        "role": ROLE,
        "model": MODEL,
        "status": "running"
    }


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.get("/ready")
async def ready():
    return {"status": "ready"}


@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):

    api_key = os.getenv("OPENAI_API_KEY")

    if not api_key:
        raise HTTPException(
            status_code=503,
            detail="OpenAI API key is not configured"
        )

    try:
        client = OpenAI(api_key=api_key)

        result = client.responses.create(
            model=MODEL,
            instructions=SYSTEM_PROMPT,
            input=request.message
        )

        return ChatResponse(
            agent=ROLE,
            model=MODEL,
            response=result.output_text
        )

    except Exception as exc:
        raise HTTPException(
            status_code=502,
            detail=f"LLM request failed: {str(exc)}"
        )
