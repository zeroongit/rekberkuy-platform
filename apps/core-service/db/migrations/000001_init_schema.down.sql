-- ============================================================================
-- 000001_init_schema.down.sql
-- Rolls back the initial schema. Tables are dropped in reverse dependency
-- order (children before parents). IF EXISTS keeps it idempotent.
-- ============================================================================

-- PHASE 5
DROP TABLE IF EXISTS disputes;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS event_vendor_allocations;
DROP TABLE IF EXISTS event_vendor_payouts;
DROP TABLE IF EXISTS event_official_details;
DROP TABLE IF EXISTS service_milestones;

-- PHASE 4
DROP TABLE IF EXISTS transaction_events;
DROP TABLE IF EXISTS transaction_services;
DROP TABLE IF EXISTS transaction_goods;

-- PHASE 3
DROP TABLE IF EXISTS rekberpay_transactions;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS kyc_submissions;
DROP TABLE IF EXISTS vendor_sub_sub_categories;
DROP TABLE IF EXISTS event_sub_sub_categories;
DROP TABLE IF EXISTS service_sub_sub_categories;
DROP TABLE IF EXISTS goods_sub_sub_categories;

-- PHASE 2
DROP TABLE IF EXISTS vendor_profiles;
DROP TABLE IF EXISTS crm_loyalty;
DROP TABLE IF EXISTS rekberpay_wallets;
DROP TABLE IF EXISTS vendor_sub_categories;
DROP TABLE IF EXISTS event_sub_categories;
DROP TABLE IF EXISTS service_sub_categories;
DROP TABLE IF EXISTS goods_sub_categories;

-- PHASE 1
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS platform_finances;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS vendor_category_models;
DROP TABLE IF EXISTS event_categories;
DROP TABLE IF EXISTS service_categories;
DROP TABLE IF EXISTS goods_categories;
