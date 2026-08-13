# Application Flow & Module Guide

This document is the **technical reference** for the RekberKuy platform: the actors, end-to-end business flows, domain state machines, money flow, and the backend module map that implements each capability. For everyday user-facing instructions, see [general-user-guide.md](./general-user-guide.md).

---

## Actors

| Actor | Description |
|-------|-------------|
| **Buyer** | User who buys goods, orders services, or purchases event tickets |
| **Seller** | User who sells physical or digital goods |
| **Service Provider** | User who offers professional / freelance services |
| **Event Organizer (EO)** | User who creates and manages events |
| **Vendor** | Provider of event support services (catering, decoration, photographer, sound system, etc.) |
| **Admin** | Internal RekberKuy team that manages the platform and mediates disputes |
| **System** | Automated processes (timers, notifications, relayer, workers) |

---

## Universal Transaction State Machine

All three business lines (Goods, Services, Events) share one rigid state machine enforced at the database level (`domain.TransactionStatus`):

```
WAITING_PAYMENT ──(payment confirmed)──► FUNDS_LOCKED ──(confirm receipt / milestone done / auto-release timeout)──► RELEASED  (success)
                                              │
                                     (dispute / complaint)
                                              ▼
                                             DISPUTED ──(admin resolution)──► REFUNDED  (funds returned)
 ```
 ---
 
## AI-Assisted Admin Verification — Current Status

`backend-ai` is a **scoring-only** service — it never makes a decision. For every capability it supports, `core-service` (or a human Admin) is the sole decision-maker: `backend-ai` returns a number, `core-service` compares it against its own threshold.

| Capability | Status | Notes |
|---|---|---|
| **Fraud scoring** | ✅ Implemented & wired | Called before every fund release: `ReleaseFundsGoods`, `ReleaseMilestoneServices`, `ReleaseFundsEvents`, and `ProcessEventVendorPayout` (external vendor payouts). |
| **KYC verification** | ❌ Not wired | `backend-ai` exposes `POST /api/v1/kyc/verify`, but `kyc_usecase.go` never calls it — `SubmitUserKYC` only persists the submission as `PENDING`. There is no admin approve/reject endpoint yet, so `KYCApproved` / `KYCRejected` are unreachable states. |
| **Dispute resolution** | 🚫 Deliberately not AI-assisted | By design, per [ADR-0003](./adr/0003-admin-mediated-dispute-resolution-ai-deferred.md): an Admin sets the binding `Outcome` (`REFUND_BUYER` / `RELEASE_TO_SELLER`) manually. `dispute_usecase.go` never calls `backend-ai`. |

### Verification chain (fraud scoring — the only capability actually wired end-to-end)

Every fund release runs this exact sequence (see e.g. `TransactionGoodsUsecase.ReleaseFundsGoods`):

```
core-service                    backend-ai                  core-service                blockchain
     │                               │                            │                          │
     ├── POST /api/v1/fraud/score ──►│                            │                          │
     │                               ├── Groq scoring ────────────│                          │
     │◄── { score, reason } ─────────┤                            │                          │
     ├── compute isSafe (score ≥ FRAUD_UNSAFE_THRESHOLD) ─────────►│                          │
     │                                                             │                          │
     │        isSafe == false  ──► release REFUSED, nothing moves, blockchain untouched        │
     │        isSafe == true   ──► ACID mutation (wallet credit + status RELEASED, in Postgres) │
     │                                                             ├── logTransaction() ──────►│
     │                                                             │   (best-effort; audit hash only, no PII)
```

Two properties worth calling out:

- **`backend-ai` never decides.** It cannot flag a transaction unsafe on its own; `isSafe` is computed inside `core-service` from `cfg.AI.UnsafeThreshold`. If `backend-ai` is unreachable, `core-service`'s own `FRAUD_FAIL_OPEN` policy decides (default `false` — fail-closed, refuse the release).
- **Blockchain is a downstream side-effect, not a payment step.** Money already moved in Postgres (wallet balances, transaction status) *before* `logTransaction()` is called. The on-chain call is a best-effort, immutable receipt of a transaction that already happened — never a payment method and never a precondition for the money to move.

---

| Status | Meaning |
|--------|---------|
| `WAITING_PAYMENT` | Transaction created; buyer has not yet paid |
| `FUNDS_LOCKED` | Buyer funds are held in escrow |
| `RELEASED` | Funds released to the seller/provider/vendor (terminal success) |
| `DISPUTED` | A complaint froze the funds pending mediation |
| `REFUNDED` | Funds returned to the buyer (terminal) |

Wallet-level ledger entries use their own status set (`PENDING`, `SUCCESS`, `FAILED`) in `domain.WalletTxStatus`, and event vendor payouts use `PENDING` / `APPROVED` / `PENDING_DISBURSEMENT`.

---

## Money Flow

```
Buyer Top-Up (Midtrans)
       │
       ▼
 Buyer Wallet (RekberPay)
       │
       │ (on transaction creation)
       ▼
 Escrow (FUNDS_LOCKED) ─────► Platform Fee (deducted on RELEASED)
       │
       │ (after confirmation / settlement)
       ▼
 Seller / Provider / EO / Vendor Wallet
       │
       │ (on withdrawal)
       ▼
 User Bank Account (Midtrans Disbursement, T+1)
```

All money is purely **IDR (Rupiah)**. The platform never holds crypto and never asks users for crypto addresses. The only on-chain activity is a **gasless audit log** (hash + amount) — no funds ever live on-chain.

---

## Goods Transaction Flow

```
Buyer                    System                   Seller
   │                        │                        │
   ├── Create Transaction ─►│                        │
   │                        ├── Deduct Balance ──────│
   │                        ├── Funds to Escrow ─────│
   │                        ├── Notify Seller ───────►│
   │                        │                        │
   │                        │◄── Confirm Shipping ───┤
   │                        ├── Update Status ───────│
   │◄── Shipped Notif ──────┤                        │
   │                        │                        │
   ├── Confirm Receipt ────►│                        │
   │                        ├── Fraud Check ─────────│
   │                        ├── Release Escrow ─────►│
   │                        ├── Log Blockchain ──────│
   │                        ├── Notify Seller ───────►│
   │◄── Completed Notif ────┤                        │
```

**Key behaviors:**
- Funds are locked the moment payment is confirmed (`ConfirmPaymentGoods`).
- If the buyer does not confirm receipt within the auto-confirm deadline, the **Auto-Release Worker** releases funds to the seller automatically.
- Before release, a **fraud-scoring** check runs; high-risk transactions are held for review.
- After release, an async **gasless audit log** is written to Avalanche and the tx hash is persisted.

---

## Services Transaction Flow (Milestone-Based)

```
Client                       System                    Service Provider
   │                            │                            │
   ├── Create Service Tx ──────►│                            │
   │                            ├── Lock Total Funds ────────│
   │                            ├── Activate Milestone 1 ────►│
   │                            │                            │
   │                            │◄── Mark Milestone Done ────┤
   │                            ├── Notify Client ───────────│
   │◄── Review Notif ───────────┤                            │
   ├── Confirm Milestone ──────►│                            │
   │                            ├── Fraud Check ─────────────│
   │                            ├── Release Milestone ──────►│
   │                            ├── Next Milestone ──────────│
   │                            │   (repeat until last)      │
   │                            ├── Log Blockchain ──────────│
```

**Key behaviors:**
- The **full contract value** is locked up front; it is released **per milestone** as the client confirms each deliverable.
- If the client does not respond within the review window, the milestone funds are released automatically.
- Each milestone release is fraud-checked and audit-logged.

---

## Event Transaction Flow

```
EO                       System                Participant / Vendor
   │                        │                        │
   ├── Create Event ───────►│                        │
   ├── Setup Vendor ────────►│                        │
   │                        ├── Publish Event ───────►│ (Participant)
   │                        │◄── Buy Ticket ─────────┤
   │                        ├── Ticket Funds Escrow ─│
   │                        │                        │
   ├── Vendor Contract ─────►│                        │
   │                        ├── Vendor Funds Escrow ─►│ (Vendor)
   │                        │◄── Work Completed ─────┤
   ├── Confirm Vendor ──────►│                        │
   │                        ├── Release to Vendor ───►│ (internal: wallet / external: Midtrans)
   │                        │                        │
   ├── Event Completed ─────►│                        │
   │                        ├── Event Audit (finance_calculator)
   │                        ├── Platform Fee ────────│
   │◄── EO Bonus + Mgmt Fee ┤                        │
   │                        ├── Auto-Refund Surplus ─►│ (Participants)
   │                        ├── Log Blockchain ──────│
```

**Key behaviors:**
- The platform locks a **5% upfront fee** on the total event budget; vendors are payable up to 95%.
- After vendor settlement, any **surplus** is split between an EO bonus (tier-based) and an auto-refund to participants.
- **Vendor payouts are hybrid**: internal vendors (platform users) are credited to their RekberPay wallet; external vendors are marked `PENDING_DISBURSEMENT` and paid via Midtrans disbursement.

---

## Vendor Marketplace Flow

```
Vendor                    System                       EO
   │                        │                           │
   ├── Register Vendor ────►│                           │
   │                        ├── Admin Review ───────────│
   │◄── Verified Badge ─────┤                           │
   │                        ├── List in Marketplace ────│
   │                        │                           │
   │                        │◄── Browse / Filter ───────┤
   │                        │◄── Create Contract ───────┤
   │                        ├── Vendor Funds Escrow ────│
   │◄── Contract Notif ─────┤                           │
```

---

## Backend Module Map

The backend (`apps/core-service/`) follows **Clean Architecture**: `delivery/handlers` → `usecase` → `repository` → `domain`. Outer layers may only depend on inner layers.

| Capability | Domain entity / interface | Usecase | Repository | Handler |
|------------|---------------------------|---------|------------|---------|
| Goods escrow | `transaction.go` (`Transaction`, `TransactionGoods`) | `transaction_goods_usecase.go` | `transaction_repository.go` | `transaction_goods_handler.go` |
| Services escrow (milestones) | `transaction.go` (`TransactionServices`, `ServiceMilestone`) | `transaction_services_usecase.go` | `transaction_repository.go` | `transaction_services_handler.go` |
| Event escrow + vendor payout | `transaction.go` (`TransactionEvents`, `EventVendorPayout`, `EventVendorAllocation`) | `transaction_events_usecase.go` | `transaction_repository.go` + `wallet_repository.go` | `transaction_events_handler.go` |
| Wallet & ledger | `wallet.go` (`RekberPayWallet`, `RekberPayTransaction`) | `wallet_usecase.go`, `user_usecase.go` (TopUp/ConfirmTopUp) | `wallet_repository.go` | `wallet_handler.go`, `midtrans_webhook_handler.go` |
| Finance & fees | `finance.go` (`PlatformFinance`, `EventAuditResult`) | `finance_calculator.go` | `finance_repository.go` | — |
| CRM Loyalty Tiering | `profile.go` (`CRMLoyalty`) | `finance_calculator.go` (tier eval) | `wallet_repository.go` | — |
| Users | `profile.go` (`UserProfile`) | `user_usecase.go` | `user_repository.go` | `user_handlers.go` |
| KYC | `profile.go` (`KYCSubmission`) | `kyc_usecase.go` | `kyc_repository.go` | `kyc_handler.go` |
| Vendors | `profile.go` (`VendorProfile`) | `vendor_usecase.go` | `vendor_repository.go` | `vendor_handler.go` |
| Categories (3-tier taxonomy) | `category.go` | — (seeded) | — | — |

### Cross-cutting infrastructure

| Concern | Port (domain) | Adapter | Wired in |
|---------|---------------|---------|----------|
| Transactional boundary (ACID) | `UnitOfWork` (`unit_of_work.go`) | `repository/unit_of_work.go` | every funds-mutating usecase |
| Fraud scoring | `FraudClient` (`services.go`) | `fraud/client.go` (HTTP → backend-ai) | release flows (before payout) |
| Gasless audit log | `Relayer` (`services.go`) | `relayer/relayer.go` (go-ethereum → Avalanche) | release flows (after commit) |
| Payment gateway | `MidtransClient` (`services.go`) | `midtrans/client.go` (Snap API) | wallet top-up + webhook |
| Idempotency (anti double-spend) | `IdempotencyRepository` (`wallet.go`) | `idempotency_repository.go` / `idempotency_redis.go` | `idempotency_middleware.go` |
| Auth (JWT + RBAC) | `auth.go` (`JWTCustomClaims`) | — | `auth_middleware.go` |

### Background workers (`internal/worker/`)

| Worker | Trigger | Responsibility |
|--------|---------|----------------|
| `auto_release_worker.go` | Every 1 hour | Releases escrow for goods transactions past their auto-confirm deadline |
| `crm_worker.go` | 1st of each month | Re-evaluates every user's CRM Loyalty Tier (Goods / Services / Events domain) |

### External services

| Service | Stack | Role |
|---------|-------|------|
| `backend-ai/` | Python 3.11 / FastAPI / Groq | Fraud risk scoring (called by `fraud/client.go`). Also exposes a KYC verification-scoring endpoint that is not yet called by core-service — see [AI-Assisted Admin Verification](#ai-assisted-admin-verification--current-status) below. |
| `blockchain/` | Solidity / Hardhat v3 / Avalanche | Audit-log smart contract `TransactionLogger` (called by `relayer/relayer.go`) |

---

## Fee & Settlement Rules

Enforced by `usecase/finance_calculator.go` (Goods/Services) and `CalculateEventAudit` (Events).

### Goods & Services — seller commission (CRM tier-based)

| Seller CRM Tier | Platform Commission (of released amount) |
|-----------------|------------------------------------------|
| NEWBIE | 10% |
| SILVER | 6% |
| GOLD | 3% (promo) |

### Event line — platform fee & EO surplus

- **Platform fee:** 5% upfront of the total locked event budget. Vendors are payable up to 95% of the locked funds; the remainder is the efficiency surplus.
- **EO surplus bonus** (tier-based share of the surplus):

| Event Fund Scale | EO NEWBIE | EO SILVER | EO GOLD |
|------------------|-----------|-----------|---------|
| Micro Event (< Rp 10.000.000) | 5% | 10% | 15% |
| Mega Event (> Rp 10.000.000) | 2% | 4% | 8% |

- **Surplus distribution:**
  - Micro surplus (≤ Rp 500.000): 100% goes to the EO's wallet as a performance bonus.
  - Macro surplus (> Rp 500.000): split between the EO bonus (tier % above) and an auto-refund to each participant's RekberPay balance.

### Buyer protection fee (charged upfront)

| Seller Tier | Buyer Protection Fee |
|-------------|----------------------|
| GOLD | 8% of base amount |
| SILVER | 4% of base amount |
| NEWBIE | Rp 2.500 via RekberPay · Rp 5.000 otherwise |

### Withdrawal fee

- **Rp 7.500** flat to a bank account (includes interbank clearing via the payment gateway, T+1).

---

*This document is maintained alongside the codebase. When a flow or module changes, update the corresponding section here.*
