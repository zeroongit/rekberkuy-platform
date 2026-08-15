-- 000002: KYC AI-assisted verification reference columns.
-- backend-ai (Groq vision) returns a confidence score + reason when a user
-- submits KYC documents. The score is a REFERENCE ONLY for the reviewing
-- admin — the decision (approve/reject) always belongs to core-service's
-- admin flow, never to the AI. Columns are nullable because the AI service
-- may be unreachable at submission time; the submission stays PENDING
-- regardless, so the admin can still review the raw documents.

ALTER TABLE kyc_submissions
    ADD COLUMN ai_score DOUBLE PRECISION,
    ADD COLUMN ai_reason TEXT;
