-- 000003: Wallet withdrawal requests.
-- The wallet balance is debited (amount + WithdrawFeeToUser) at request time;
-- the actual bank transfer completes out-of-band (same stance as ADR-0002 for
-- external vendor disbursements) and an admin confirms completion, flipping
-- the status to PAID.

CREATE TABLE withdrawals (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID         NOT NULL REFERENCES user_profiles(id),
    amount         BIGINT       NOT NULL CHECK (amount > 0),
    fee            BIGINT       NOT NULL,
    bank_name      VARCHAR(100) NOT NULL,
    account_number VARCHAR(100) NOT NULL,
    account_holder VARCHAR(255) NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    processed_by   UUID         REFERENCES user_profiles(id),
    processed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
