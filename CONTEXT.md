# RekberKuy

An escrow (joint-account) platform for **Goods**, **Services**, and **Events**. Funds are locked in IDR, released on completion, and every completed transaction is recorded as a gasless audit-log entry on Avalanche. This glossary defines the ubiquitous language used across the backend and docs.

## Language

### Core money concepts

**Transaction**:
The universal escrow record that locks funds between two parties for a goods, services, or event deal (`domain.Transaction`). Has one shared state machine: `WAITING_PAYMENT → FUNDS_LOCKED → RELEASED`, with `DISPUTED → REFUNDED` as the complaint branch.
_Avoid_: order, payment, deal.

**Ledger Entry**:
A single line on a RekberPay wallet's balance — a top-up, payment, refund, or withdrawal (`domain.RekberPayTransaction`). A wallet has many ledger entries; a Transaction has none.
_Avoid_: transaction (that word is reserved for the escrow record), wallet tx.

**Release**:
Move locked escrow funds to the counterparty's RekberPay wallet (seller / service provider / EO) on successful completion.
_Avoid_: disburse, pay out, settle, transfer.

**Disburse**:
Move funds out of the platform to an external bank account — an external event vendor's payout or a user withdrawal. The platform is no longer holding the money afterwards.
_Avoid_: release, withdraw (a _withdrawal_ is a user-initiated disbursement to their own bank).

**Refund**:
Return locked escrow funds to the buyer's RekberPay wallet (the outcome of a dispute resolved in the buyer's favour).
_Avoid_: reversal, chargeback, return.

### Roles & actors

A role-in-a-transaction is **not** the same thing as a `UserRole`. Any user can be a buyer in one deal and a seller in another.

**Buyer / Seller**:
The two parties on a Transaction (`buyer_id` / `seller_id`). Roles-in-a-transaction, not UserRoles.
_Avoid_: customer, merchant, client, provider (those map to specific UserRoles below).

**Verified Merchant** (`UserRole VERIFIED_MERCHANT`):
A user who has passed KYC to sell; their CRM Loyalty tier sets their commission rate.
_Avoid_: seller (that's role-in-a-transaction), vendor.

**Vendor** (`UserRole VERIFIED_VENDOR`):
A registered event support provider (catering, sound system, decor, venue) with a `VendorProfile`.
_Avoid_: merchant, seller.

**Event Organizer / EO**:
The user who creates an event Transaction and manages its vendors. Reuses the `VendorProfile` shape (`EOProfile`).
_Avoid_: organizer, host.

**External Vendor**:
An event vendor without a RekberPay account — paid by bank **disbursement** rather than internal wallet credit (`EventVendorPayout.VendorUserID == nil`).
_Avoid_: offline vendor, third-party vendor.

**Admin** (`UserRole ADMIN`):
Internal RekberKuy staff who mediate disputes, approve KYC, and mark external disbursements complete.

### Disputes

**Dispute**:
A formal complaint raised by either party against a Transaction while its funds are locked, freezing the escrow pending admin mediation. One Dispute per Transaction.
_Avoid_: complaint, ticket, case.

**Dispute Outcome**:
The mediator's binding decision on a resolved Dispute — either `REFUND_BUYER` or `RELEASE_TO_SELLER`. Determines whether the Transaction ends in `REFUNDED` or `RELEASED`.
_Avoid_: verdict, resolution.

### Reputation

**Review**:
A 1–5 rating with optional comment, left by the buyer on the counterparty after a Transaction is `RELEASED`. One per `(transaction, reviewer)`. Aggregated by the CRM worker into the merchant's average rating.
_Avoid_: feedback, rating (a _rating_ is the numeric score inside a Review).

**CRM Loyalty Tier**:
A merchant's rolling standing (`NEWBIE / SILVER / GOLD`) recomputed monthly from GMV, completed deals, and review average. Sets the commission and buyer-protection fees.
