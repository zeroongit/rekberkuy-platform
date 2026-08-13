-- ============================================================================
-- 000001_init_schema.up.sql
-- RekberKuy Core Service — initial schema (PostgreSQL / Supabase).
--
-- Derived from the entity structs in internal/domain/.
-- Tables are created in dependency order (parents before children).
-- Foreign keys use:
--   ON DELETE CASCADE  -> where the domain struct tags constraint:OnDelete:CASCADE
--   ON DELETE NO ACTION -> default for all other relations (PostgreSQL default)
--
-- IMPORTANT naming notes (must match the repository raw-SQL queries):
--   * RekberPayWallet     -> rekberpay_wallets      (not "rekber_pay_wallets")
--   * RekberPayTransaction-> rekberpay_transactions (not "rekber_pay_transactions")
--   * CRMLoyalty          -> crm_loyalty            (singular)
-- ============================================================================

-- gen_random_uuid(): built into PostgreSQL 13+, extension guard for older setups.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- PHASE 1 — independent master / root tables
-- ============================================================================

CREATE TABLE goods_categories (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE service_categories (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE event_categories (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_category_models (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    id            UUID         PRIMARY KEY,
    username      VARCHAR(255) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'USER',
    phone_number  VARCHAR(50),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Singleton row (id = 'GLOBAL_FINANCE_ID') holding the platform cash summary.
CREATE TABLE platform_finances (
    id                   VARCHAR(64)  PRIMARY KEY,
    total_escrow_balance BIGINT       NOT NULL DEFAULT 0,
    total_revenue        BIGINT       NOT NULL DEFAULT 0,
    total_midtrans_fees  BIGINT       NOT NULL DEFAULT 0,
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE idempotency_records (
    id              VARCHAR(255) PRIMARY KEY,
    request_path    VARCHAR(255) NOT NULL,
    response_body   BYTEA,
    response_status INTEGER      NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- PHASE 2 — depend on phase 1 root tables
-- ============================================================================

CREATE TABLE goods_sub_categories (
    id          BIGSERIAL    PRIMARY KEY,
    category_id BIGINT       NOT NULL REFERENCES goods_categories(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE service_sub_categories (
    id          BIGSERIAL    PRIMARY KEY,
    category_id BIGINT       NOT NULL REFERENCES service_categories(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE event_sub_categories (
    id          BIGSERIAL    PRIMARY KEY,
    category_id BIGINT       NOT NULL REFERENCES event_categories(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_sub_categories (
    id          BIGSERIAL    PRIMARY KEY,
    category_id BIGINT       NOT NULL REFERENCES vendor_category_models(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE rekberpay_wallets (
    user_id    UUID        PRIMARY KEY REFERENCES user_profiles(id) ON DELETE CASCADE,
    balance    BIGINT      NOT NULL DEFAULT 0,
    is_frozen  BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE crm_loyalty (
    user_id                    UUID         PRIMARY KEY REFERENCES user_profiles(id) ON DELETE CASCADE,
    total_points               BIGINT       NOT NULL DEFAULT 0,
    current_tier               VARCHAR(50)  NOT NULL DEFAULT 'BRONZE',
    total_spent_fiat           BIGINT       NOT NULL DEFAULT 0,
    rolling_3_month_gmv        BIGINT       NOT NULL DEFAULT 0,
    current_month_gmv          BIGINT       NOT NULL DEFAULT 0,
    max_item_price_sold        BIGINT       NOT NULL DEFAULT 0,
    total_completed_services   INTEGER      NOT NULL DEFAULT 0,
    total_completed_events     INTEGER      NOT NULL DEFAULT 0,
    consecutive_failed_months  INTEGER      NOT NULL DEFAULT 0,
    tier_evaluation_started_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_month_evaluated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                 TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_profiles (
    vendor_id     UUID         PRIMARY KEY REFERENCES user_profiles(id) ON DELETE CASCADE,
    business_name VARCHAR(255) NOT NULL,
    category      VARCHAR(100) NOT NULL,
    is_verified   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- PHASE 3 — depend on phase 1-2 tables
-- ============================================================================

CREATE TABLE goods_sub_sub_categories (
    id              BIGSERIAL    PRIMARY KEY,
    sub_category_id BIGINT       NOT NULL REFERENCES goods_sub_categories(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE service_sub_sub_categories (
    id              BIGSERIAL    PRIMARY KEY,
    sub_category_id BIGINT       NOT NULL REFERENCES service_sub_categories(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE event_sub_sub_categories (
    id              BIGSERIAL    PRIMARY KEY,
    sub_category_id BIGINT       NOT NULL REFERENCES event_sub_categories(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_sub_sub_categories (
    id              BIGSERIAL    PRIMARY KEY,
    sub_category_id BIGINT       NOT NULL REFERENCES vendor_sub_categories(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE kyc_submissions (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID         NOT NULL UNIQUE REFERENCES user_profiles(id) ON DELETE CASCADE,
    target_role    VARCHAR(50)  NOT NULL,
    id_card_number VARCHAR(50)  NOT NULL UNIQUE,
    id_card_url    TEXT         NOT NULL,
    selfie_url     TEXT         NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    admin_notes    TEXT,
    reviewed_by    UUID         REFERENCES user_profiles(id),
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id             UUID         NOT NULL REFERENCES user_profiles(id),
    seller_id            UUID         NOT NULL REFERENCES user_profiles(id),
    type                 VARCHAR(50)  NOT NULL,
    status               VARCHAR(50)  NOT NULL DEFAULT 'WAITING_PAYMENT',
    amount_base          BIGINT       NOT NULL,
    shipping_fee         BIGINT       NOT NULL DEFAULT 0,
    service_fee          BIGINT       NOT NULL,
    midtrans_fee         BIGINT       NOT NULL,
    amount_gross         BIGINT       NOT NULL,
    amount_net           BIGINT       NOT NULL,
    midtrans_order_id    VARCHAR(255) NOT NULL UNIQUE,
    idempotency_key      VARCHAR(255) NOT NULL UNIQUE,
    payment_method       VARCHAR(100) NOT NULL,
    blockchain_tx_hash   VARCHAR(255) UNIQUE,
    blockchain_logged_at TIMESTAMPTZ,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_buyer_id  ON transactions(buyer_id);
CREATE INDEX idx_transactions_seller_id ON transactions(seller_id);

-- wallet_id references rekberpay_wallets.user_id (the wallet's primary key).
CREATE TABLE rekberpay_transactions (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id                UUID         NOT NULL REFERENCES rekberpay_wallets(user_id) ON DELETE CASCADE,
    type                     VARCHAR(50)  NOT NULL,
    status                   VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    amount                   BIGINT       NOT NULL,
    admin_fee                BIGINT       NOT NULL DEFAULT 0,
    platform_net_profit      BIGINT       NOT NULL DEFAULT 0,
    reference_transaction_id UUID,
    midtrans_topup_id        VARCHAR(255),
    description              TEXT,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rekberpay_transactions_wallet_id ON rekberpay_transactions(wallet_id);

-- ============================================================================
-- PHASE 4 — transaction detail containers (depend on phase 3)
-- ============================================================================

CREATE TABLE transaction_goods (
    transaction_id           UUID         PRIMARY KEY REFERENCES transactions(id) ON DELETE CASCADE,
    sub_sub_category_id      BIGINT       NOT NULL REFERENCES goods_sub_sub_categories(id),
    shipping_courier         VARCHAR(100) NOT NULL,
    shipping_tracking_number VARCHAR(255),
    shipping_address         TEXT         NOT NULL,
    auto_confirm_deadline    TIMESTAMPTZ  NOT NULL
);

CREATE TABLE transaction_services (
    transaction_id      UUID         PRIMARY KEY REFERENCES transactions(id) ON DELETE CASCADE,
    sub_sub_category_id BIGINT       NOT NULL REFERENCES service_sub_sub_categories(id),
    project_deadline    TIMESTAMPTZ  NOT NULL,
    brief_description   TEXT         NOT NULL
);

CREATE TABLE transaction_events (
    transaction_id        UUID         PRIMARY KEY REFERENCES transactions(id) ON DELETE CASCADE,
    sub_sub_category_id   BIGINT       NOT NULL REFERENCES event_sub_sub_categories(id),
    event_name            VARCHAR(255) NOT NULL,
    event_start_time      TIMESTAMPTZ  NOT NULL,
    event_end_time        TIMESTAMPTZ  NOT NULL,
    ticket_quantity_total INTEGER      NOT NULL DEFAULT 0
);

-- ============================================================================
-- PHASE 5 — depend on phase 4 detail containers
-- ============================================================================

CREATE TABLE service_milestones (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id  UUID         NOT NULL REFERENCES transaction_services(transaction_id) ON DELETE CASCADE,
    milestone_index INTEGER      NOT NULL,
    title           VARCHAR(255) NOT NULL,
    amount          BIGINT       NOT NULL,
    status          VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    released_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_milestones_transaction_id ON service_milestones(transaction_id);

CREATE TABLE event_official_details (
    transaction_id UUID         PRIMARY KEY REFERENCES transaction_events(transaction_id) ON DELETE CASCADE,
    organizer_id   UUID         NOT NULL REFERENCES vendor_profiles(vendor_id),
    management_fee BIGINT       NOT NULL,
    approved_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE event_vendor_payouts (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id           UUID         NOT NULL REFERENCES transaction_events(transaction_id) ON DELETE CASCADE,
    vendor_user_id           UUID         REFERENCES user_profiles(id),
    vendor_name              VARCHAR(255) NOT NULL,
    vendor_bank_name         VARCHAR(100) NOT NULL,
    vendor_account_number    VARCHAR(100) NOT NULL,
    amount_requested         BIGINT       NOT NULL,
    expense_description      TEXT         NOT NULL,
    invoice_file_url         TEXT         NOT NULL,
    payout_phase             VARCHAR(100) NOT NULL DEFAULT 'FINAL_SETTLEMENT',
    status                   VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    is_disbursed_by_midtrans BOOLEAN      NOT NULL DEFAULT FALSE,
    disbursed_at             TIMESTAMPTZ,
    reviewed_by              UUID         REFERENCES user_profiles(id),
    reviewed_at              TIMESTAMPTZ,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_vendor_payouts_transaction_id ON event_vendor_payouts(transaction_id);

CREATE TABLE event_vendor_allocations (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id     UUID         NOT NULL REFERENCES transaction_events(transaction_id) ON DELETE CASCADE,
    vendor_id          UUID         NOT NULL REFERENCES vendor_profiles(vendor_id),
    allocated_amount   BIGINT       NOT NULL,
    actual_paid_amount BIGINT       DEFAULT 0,
    status             VARCHAR(50)  DEFAULT 'PLEDGED',
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_vendor_allocations_transaction_id ON event_vendor_allocations(transaction_id);

CREATE TABLE reviews (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID         NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    reviewer_id    UUID         NOT NULL REFERENCES user_profiles(id),
    reviewee_id    UUID         NOT NULL REFERENCES user_profiles(id),
    rating         INTEGER      NOT NULL,
    comment        TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_review_tx_reviewer ON reviews(transaction_id, reviewer_id);
CREATE INDEX idx_reviews_reviewee_id ON reviews(reviewee_id);

CREATE TABLE disputes (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id     UUID         NOT NULL UNIQUE REFERENCES transactions(id) ON DELETE CASCADE,
    raised_by          UUID         NOT NULL REFERENCES user_profiles(id),
    target_party_id    UUID         REFERENCES user_profiles(id),
    mediator_id        UUID         REFERENCES user_profiles(id),
    reason             TEXT         NOT NULL,
    evidence_url       TEXT,
    status             VARCHAR(50)  NOT NULL DEFAULT 'OPEN',
    outcome            VARCHAR(50),
    is_resolved        BOOLEAN      NOT NULL DEFAULT FALSE,
    resolution_summary TEXT,
    resolved_at        TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
