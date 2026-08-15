ALTER TABLE kyc_submissions
    DROP COLUMN IF EXISTS ai_score,
    DROP COLUMN IF EXISTS ai_reason;
