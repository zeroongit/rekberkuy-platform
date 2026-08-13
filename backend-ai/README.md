# 🤖 RekberKuy — Backend AI

Python 3.11 / FastAPI microservice providing **KYC + fraud-risk scoring** via the [Groq](https://console.groq.com/) SDK. This is the verification adapter target of the Go `core-service`.

> **Architectural rule:** this service is **verification only — it never decides.** It returns a numeric `score` (+ `reason`). The Go `core-service` owns every decision: it computes `is_safe` from the fraud score using its own `FRAUD_UNSAFE_THRESHOLD`, and an Admin / core-service approves KYC. On any internal failure this service returns HTTP 503 so core-service applies its own fail-open / fail-closed policy.

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/api/v1/fraud/score` | Score a transaction's fraud risk (canonical) |
| `POST` | `/fraud/analyze` | Legacy alias of the above (kept for backward compatibility; `core-service` calls the canonical path) |
| `POST` | `/api/v1/kyc/verify` | Score a KYC submission (ID card vs. selfie) |
| `GET`  | `/healthz` | Liveness probe (+ reports whether Groq is configured) |
| `GET`  | `/docs` | Swagger UI |

### Contracts

**Fraud** — returns a score only; `core-service` decides `is_safe`:
```json
// POST /api/v1/fraud/score   (and /fraud/analyze)
// request
{ "user_id": "uuid", "amount": 1500000 }
// response  (NO is_safe here — core-service computes it from score + its FRAUD_UNSAFE_THRESHOLD)
{ "score": 0.18, "reason": "amount within normal range" }
```

**KYC** — returns a confidence score only; approval is an Admin / core-service decision:
```json
// POST /api/v1/kyc/verify
// request
{ "user_id": "uuid", "id_card_url": "https://...", "selfie_url": "https://...", "target_role": "VERIFIED_MERCHANT" }
// response  (NO is_verified here — the backend/admin decides approval)
{ "score": 0.82, "reason": "face matches ID" }
```

## Run locally

```bash
cd backend-ai
python3.11 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env          # then put your GROQ_API_KEY in .env
python main.py                # serves on :8081 (matches core-service AI_SERVICE_URL default)
```

Or with uvicorn directly: `uvicorn main:app --reload --port 8081`.

**Groq is optional at runtime.** If `GROQ_API_KEY` is unset the service still boots and returns deterministic fallback scores (`/healthz` reports `groq_configured: false`), so core-service development is never blocked.

## Configuration

See `.env.example`. The service keeps NO decision thresholds or policies — those live in `core-service`. Here it only configures the Groq models and the server port.

## Notes & limitations
- Fraud uses a text model (inputs are `user_id` + `amount` only); KYC uses a vision model comparing the two images.
- On a Groq call failure the endpoint returns HTTP 503; `core-service` then applies its `FRAUD_FAIL_OPEN` policy (fail-closed = refuse release, the default).
- The legacy `/fraud/analyze` alias exists so any caller still using it keeps working. `core-service` calls the canonical `/api/v1/fraud/score`.
