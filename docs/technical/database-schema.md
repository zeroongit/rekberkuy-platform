# Database Schema & ERD — RekberKuy Platform

## 1. Overview
Database utama RekberKuy menggunakan **PostgreSQL (Supabase)** dengan dukungan transaksi ACID, indeks performa tinggi, dan pembatasan constraint ketat untuk menjaga integritas saldo dompet serta status escrow.

---

## 2. Relational Schema (Core Tables)

### A. `user_profiles`
Menyimpan profil pengguna, kredensial autentikasi, role, dan status KYC.
- `id` (UUID, PK)
- `email` (VARCHAR, Unique)
- `password_hash` (VARCHAR)
- `full_name` (VARCHAR)
- `phone` (VARCHAR)
- `role` (ENUM: `USER`, `VERIFIED_MERCHANT`, `VERIFIED_VENDOR`, `EVENT_ORGANIZER`, `ADMIN`)
- `kyc_status` (ENUM: `PENDING`, `APPROVED`, `REJECTED`)
- `created_at`, `updated_at`

### B. `rekberpay_wallets`
Menyimpan saldo dompet pengguna (tersedia dan tertahan).
- `id` (UUID, PK)
- `user_id` (UUID, FK to `user_profiles`)
- `balance` (DECIMAL(18,2), Default 0.00)
- `escrow_balance` (DECIMAL(18,2), Default 0.00)
- `created_at`, `updated_at`

### C. `rekberpay_transactions` (Ledger)
Mencatat seluruh mutasi dan riwayat finansial dompet secara ganda (double-entry).
- `id` (UUID, PK)
- `wallet_id` (UUID, FK to `rekberpay_wallets`)
- `amount` (DECIMAL(18,2))
- `type` (ENUM: `TOPUP`, `WITHDRAW`, `ESCROW_LOCK`, `ESCROW_RELEASE`, `REFUND`, `FEE`)
- `status` (ENUM: `PENDING`, `SUCCESS`, `FAILED`)
- `reference_id` (VARCHAR)
- `created_at`

### D. `transactions` (Universal Escrow Entity)
Entitas utama transaksi lintas domain (Goods, Services, Events).
- `id` (UUID, PK)
- `domain_type` (ENUM: `GOODS`, `SERVICES`, `EVENTS`)
- `buyer_id` (UUID, FK)
- `seller_id` (UUID, FK)
- `amount` (DECIMAL(18,2))
- `status` (ENUM: `DRAFT`, `WAITING_PAYMENT`, `CANCELLED`, `FUNDS_LOCKED`, `RELEASED`, `DISPUTED`, `REFUNDED`)
- `blockchain_tx_hash` (VARCHAR, Nullable)
- `created_at`, `updated_at`

### E. `service_milestones`
Tahapan pencairan milestone untuk transaksi jasa.
- `id` (UUID, PK)
- `transaction_id` (UUID, FK)
- `title` (VARCHAR)
- `amount` (DECIMAL(18,2))
- `status` (ENUM: `PENDING`, `IN_PROGRESS`, `COMPLETED`, `RELEASED`)
- `deliverable_url` (VARCHAR)

### F. `disputes`
Pencatatan sengketa transaksi dan keputusan mediasi admin.
- `id` (UUID, PK)
- `transaction_id` (UUID, FK)
- `complainant_id` (UUID, FK)
- `reason` (TEXT)
- `evidence_url` (VARCHAR)
- `status` (ENUM: `OPEN`, `UNDER_REVIEW`, `RESOLVED`)
- `outcome` (ENUM: `REFUND_BUYER`, `RELEASE_TO_SELLER`, `SPLIT`)
- `admin_notes` (TEXT)
- `resolved_at`

---

## 3. Entity Relationship Diagram (ERD) Overview

```
user_profiles (1) ──── (1) rekberpay_wallets (1) ──── (N) rekberpay_transactions
     │
     ├────── (1:N) ──────► transactions (Universal Escrow)
     │                          │
     │                          ├────── (1:N) ──────► service_milestones
     │                          ├────── (1:1) ──────► disputes
     │                          └────── (1:N) ──────► event_vendor_payouts
     │
     └────── (1:1) ──────► vendor_profiles / kyc_submissions
```
