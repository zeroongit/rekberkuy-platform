"""
RekberKuy backend-ai — KYC + fraud-risk SCORING microservice (verification only).

Stack: Python 3.11 / FastAPI / Groq.

ARCHITECTURE RULE: this service is a VERIFICATION helper, NOT a decision maker.
It returns a numeric score (and a human-readable reason) only. The Go
core-service owns every decision:
  * the fraud "is_safe" verdict is computed in core-service from this score
    using core-service's own FRAUD_UNSAFE_THRESHOLD;
  * the KYC "is_verified" verdict is taken by an Admin / core-service, not here.
On any internal failure (e.g. Groq error) this service returns HTTP 503 so the
caller (core-service) applies its own fail-open / fail-closed policy.

Groq is optional at runtime: when GROQ_API_KEY is unset the service still boots
and returns a deterministic low-confidence score, so development is not blocked.
"""

from __future__ import annotations

import json
import logging
import os
import re
from typing import Any, Optional

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

load_dotenv()

logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"), format="%(asctime)s %(levelname)s %(name)s — %(message)s")
logger = logging.getLogger("backend-ai")

# ---------------------------------------------------------------------------
# Configuration (env-driven; no hardcoded secrets)
# ---------------------------------------------------------------------------
GROQ_API_KEY: Optional[str] = os.getenv("GROQ_API_KEY") or None
GROQ_TEXT_MODEL: str = os.getenv("GROQ_MODEL", "llama-3.3-70b-versatile")
GROQ_VISION_MODEL: str = os.getenv("GROQ_VISION_MODEL", "meta-llama/llama-4-scout-17b-16e-instruct")

_groq: Any = None
if GROQ_API_KEY:
    try:
        from groq import Groq

        _groq = Groq(api_key=GROQ_API_KEY)
        logger.info("Groq client initialised (text=%s, vision=%s)", GROQ_TEXT_MODEL, GROQ_VISION_MODEL)
    except Exception as exc:  # pragma: no cover — import/config error
        logger.error("Groq client init failed; will return low-confidence scores: %s", exc)


def _clamp01(value: float) -> float:
    return max(0.0, min(1.0, value))


def _extract_json(content: str) -> dict[str, Any]:
    """Best-effort parse of a JSON object out of an LLM text response."""
    if not content:
        return {}
    text = content.strip()
    fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
    if fenced:
        text = fenced.group(1)
    start, end = text.find("{"), text.rfind("}")
    if start != -1 and end != -1 and end > start:
        text = text[start : end + 1]
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        return {}


app = FastAPI(
    title="RekberKuy Backend AI",
    description="KYC + fraud-risk SCORING (verification only — core-service decides)",
    version="2.0.0",
)


# ---------------------------------------------------------------------------
# Contracts — score + reason only. No verdict fields: the backend decides.
# ---------------------------------------------------------------------------
class FraudRequest(BaseModel):
    user_id: str = Field(..., description="The transacting user id")
    amount: int = Field(..., description="Transaction amount in IDR (Rupiah)")


class FraudResponse(BaseModel):
    score: float = Field(..., description="Risk score in [0.0, 1.0] — higher is riskier")
    reason: Optional[str] = None


class KYCRequest(BaseModel):
    user_id: str
    id_card_url: str = Field(..., description="Public URL of the ID-card image")
    selfie_url: str = Field(..., description="Public URL of the selfie image")
    target_role: str = Field("VERIFIED_MERCHANT", description="Role the user is applying for")


class KYCResponse(BaseModel):
    score: float = Field(..., description="Verification confidence in [0.0, 1.0]")
    reason: Optional[str] = None


# ---------------------------------------------------------------------------
# /api/v1/fraud/score  (+ legacy /fraud/analyze alias)
# ---------------------------------------------------------------------------
_FRAUD_SYSTEM = (
    "You are a fraud-risk analyst for RekberKuy, an Indonesian escrow platform "
    "handling IDR (Rupiah). Given a user id and a transaction amount, estimate a "
    "fraud risk score. Consider: unusually large amounts, round-number/structured "
    "amounts, and that a brand-new user id is riskier. You have ONLY these two "
    "inputs, so reason heuristically. Respond ONLY with compact JSON of the form "
    '{"risk_score": <0.0-1.0>, "reason": "<short>"}'
)


def score_fraud(req: FraudRequest) -> FraudResponse:
    if _groq is None:
        logger.warning("GROQ_API_KEY not set — returning low-risk score for user=%s", req.user_id)
        return FraudResponse(score=0.1, reason="groq-not-configured-fallback")

    try:
        completion = _groq.chat.completions.create(
            model=GROQ_TEXT_MODEL,
            temperature=0.2,
            response_format={"type": "json_object"},
            messages=[
                {"role": "system", "content": _FRAUD_SYSTEM},
                {"role": "user", "content": f'user_id="{req.user_id}" amount={req.amount} (IDR)'},
            ],
        )
        data = _extract_json(completion.choices[0].message.content or "")
        score = _clamp01(float(data.get("risk_score", data.get("score", 0.5))))
        return FraudResponse(score=score, reason=str(data.get("reason", "")) or None)
    except Exception as exc:
        # Verification failed — surface a 503 so core-service applies its own
        # fail-open/fail-closed policy. This service never makes the call.
        logger.exception("fraud scoring failed for user=%s", req.user_id)
        raise HTTPException(status_code=503, detail=f"scoring-error: {type(exc).__name__}")


@app.post("/api/v1/fraud/score", response_model=FraudResponse, tags=["fraud"])
@app.post("/fraud/analyze", response_model=FraudResponse, tags=["fraud"], include_in_schema=False)
def fraud_score(req: FraudRequest) -> FraudResponse:
    """Score a transaction's fraud risk (verification only). /fraud/analyze is a legacy alias."""
    return score_fraud(req)


# ---------------------------------------------------------------------------
# /api/v1/kyc/verify
# ---------------------------------------------------------------------------
_KYC_SYSTEM = (
    "You are a KYC verifier for RekberKuy. You are given an Indonesian identity "
    "card image and a selfie image. Assess whether they appear genuine and depict "
    "the same person. Respond ONLY with compact JSON of the form "
    '{"verification_score": <0.0-1.0>, "reason": "<short>"}. '
    "If either image is missing, blank, or unreadable, return a low score."
)


def verify_kyc(req: KYCRequest) -> KYCResponse:
    if _groq is None:
        logger.warning("GROQ_API_KEY not set — returning low-confidence KYC score for user=%s", req.user_id)
        return KYCResponse(score=0.0, reason="groq-not-configured-fallback")

    try:
        completion = _groq.chat.completions.create(
            model=GROQ_VISION_MODEL,
            temperature=0.2,
            response_format={"type": "json_object"},
            messages=[
                {"role": "system", "content": _KYC_SYSTEM},
                {
                    "role": "user",
                    "content": [
                        {"type": "text", "text": f'Verify KYC for user="{req.user_id}" target_role={req.target_role}.'},
                        {"type": "image_url", "image_url": {"url": req.id_card_url}},
                        {"type": "image_url", "image_url": {"url": req.selfie_url}},
                    ],
                },
            ],
        )
        data = _extract_json(completion.choices[0].message.content or "")
        score = _clamp01(float(data.get("verification_score", data.get("score", 0.0))))
        return KYCResponse(score=score, reason=str(data.get("reason", "")) or None)
    except Exception as exc:
        logger.exception("kyc verification failed for user=%s", req.user_id)
        raise HTTPException(status_code=503, detail=f"verification-error: {type(exc).__name__}")


@app.post("/api/v1/kyc/verify", response_model=KYCResponse, tags=["kyc"])
def kyc_verify(req: KYCRequest) -> KYCResponse:
    """Score a KYC submission (ID card vs. selfie). Verification only — core-service/admin decides approval."""
    return verify_kyc(req)


# ---------------------------------------------------------------------------
# Health & root
# ---------------------------------------------------------------------------
@app.get("/healthz", tags=["meta"])
def healthz() -> dict[str, Any]:
    return {"status": "ok", "groq_configured": _groq is not None}


@app.get("/", tags=["meta"])
def root() -> dict[str, str]:
    return {"service": "rekberkuy-backend-ai", "docs": "/docs"}


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8081"))
    uvicorn.run("main:app", host=os.getenv("HOST", "0.0.0.0"), port=port, reload=os.getenv("RELOAD", "false").lower() in {"1", "true"})
