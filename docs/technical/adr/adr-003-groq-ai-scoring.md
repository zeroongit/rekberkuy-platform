# ADR-003: Groq AI as a Scoring-Only Service (Human-in-the-Loop)

## Status
ACCEPTED

## Context
Automated AI decision-making for high-stakes financial operations (such as fraud flagging, KYC approval, and dispute resolution) carries significant risks of false positives and regulatory non-compliance if an AI model makes autonomous financial decisions.

## Decision
Design the AI service (`backend-ai` running FastAPI and Groq LLaMA models) strictly as a **scoring-only verification helper**. 
- `backend-ai` evaluates risk scores or KYC confidence and returns numerical metrics and reasoning.
- `core-service` evaluates the score against strict internal thresholds (`FRAUD_UNSAFE_THRESHOLD`).
- Human Administrators retain absolute veto and decision authority for KYC approvals (`POST /api/v1/admin/kyc/:id/review`) and dispute resolution outcomes.

## Consequences
- The system eliminates autonomous AI errors in funds movement.
- Core business rules remain fully deterministic and auditable within Go usecases.
- If the AI service is unreachable, fail-open/fail-closed fallback policies in `core-service` prevent system deadlock.
