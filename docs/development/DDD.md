# Domain-Driven Design (DDD)
# RekberKuy Platform

> DDD adalah fondasi dari semua dokumen lain (EDD, BDD, TDD).
> Baca ini dulu sebelum yang lain.

---

## Ubiquitous Language

Kamus istilah yang WAJIB dipakai konsisten di seluruh codebase,
dokumentasi, dan komunikasi tim. Jangan pakai sinonim.

| Istilah | Definisi Bisnis | Jangan Pakai |
|---------|----------------|--------------|
| **Escrow** | Dana yang dikunci platform sampai kondisi terpenuhi | rekber, titipan |
| **Lock** | Proses mengunci dana pembeli ke escrow | hold, freeze (kecuali dispute) |
| **Release** | Pelepasan dana escrow ke pihak penerima | cairkan, transfer |
| **Freeze** | Dana dibekukan saat sengketa aktif | lock (bedakan konteks) |
| **Milestone** | Tahapan pembayaran dalam kontrak jasa | termin, cicilan, progress |
| **Payout Request** | Pengajuan pencairan oleh vendor ke EO | invoice request, klaim |
| **Auto-Release** | Pelepasan otomatis escrow setelah timeout | auto-confirm, otomatis cairkan |
| **Dispute** | Sengketa yang dibuka salah satu pihak | komplain, laporan |
| **Settlement** | Penyelesaian akhir transaksi | closing, selesai |
| **Verified Merchant** | Penjual yang sudah KYC terverifikasi | seller verified, penjual aktif |
| **EO** | Event Organizer — pengguna yang membuat event | panitia, penyelenggara |
| **Vendor** | Penyedia layanan event (katering, sound, dll) | supplier, mitra |
| **RekberPay** | Nama wallet internal platform | saldo, balance |
| **Gross Amount** | Nilai transaksi sebelum dipotong fee | harga kotor |
| **Net Payout** | Nilai yang diterima penerima setelah fee | harga bersih, take-home |
| **Fraud Score** | Nilai risiko transaksi dari AI (0.0–1.0) | risk level, skor bahaya |
| **KYC** | Proses verifikasi identitas pengguna | verifikasi, validasi akun |
| **Gasless** | Transaksi blockchain tanpa user bayar gas | gratis blockchain |
| **Relayer** | Backend yang kirim tx ke blockchain atas nama user | bot, agent |

---

## Bounded Contexts

Setiap bounded context adalah wilayah yang memiliki model domain sendiri
dan bahasa yang konsisten di dalamnya.

```
┌─────────────────────────────────────────────────────────────────┐
│                     RekberKuy Platform                          │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │   Identity   │  │  Transaction │  │      Financial        │ │
│  │   Context    │  │   Context    │  │      Context          │ │
│  │              │  │              │  │                       │ │
│  │ User         │  │ Escrow       │  │ Wallet (RekberPay)    │ │
│  │ Auth         │  │ Goods        │  │ Ledger                │ │
│  │ KYC          │  │ Services     │  │ Fee Calculator        │ │
│  │ Profile      │  │ Events       │  │ Disbursement          │ │
│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │    Vendor    │  │     Audit    │  │    Notification       │ │
│  │   Context    │  │   Context    │  │      Context          │ │
│  │              │  │              │  │                       │ │
│  │ Vendor       │  │ Blockchain   │  │ Email                 │ │
│  │ Marketplace  │  │ AuditLog     │  │ In-App                │ │
│  │ Category     │  │ EventRecord  │  │ Push                  │ │
│  └──────────────┘  └──────────────┘  └───────────────────────┘ │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    AI Context                            │  │
│  │         Fraud Scoring · KYC Verification (backend-ai)   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Aggregates, Entities & Value Objects

### Transaction Context

```
Transaction (Aggregate Root)
├── id: TransactionID          [Value Object]
├── type: TransactionType      [Value Object] GOODS | SERVICES | EVENTS
├── status: TransactionStatus  [Value Object] → lihat State Machine di EDD.md
├── grossAmount: Money         [Value Object] amount + currency (IDR)
├── netPayout: Money           [Value Object] grossAmount - platform fee
├── platformFee: Money         [Value Object]
├── buyerId: UserID            [Value Object]
├── sellerId: UserID           [Value Object]
├── idempotencyKey: string     [Value Object]
│
├── TransactionDetail[]        [Entity]
│   ├── itemName
│   ├── quantity
│   └── unitPrice: Money
│
└── Milestone[]                [Entity] — hanya untuk SERVICES
    ├── milestoneId
    ├── title
    ├── amount: Money
    ├── deadline: Date
    └── status: MilestoneStatus
```

### Financial Context

```
Wallet (Aggregate Root)
├── walletId: WalletID         [Value Object]
├── userId: UserID             [Value Object]
├── availableBalance: Money    [Value Object]
├── lockedBalance: Money       [Value Object] — dana di escrow
│
└── WalletLedger[]             [Entity]
    ├── ledgerId
    ├── type: LedgerType       CREDIT | DEBIT
    ├── amount: Money
    ├── reference: string      transaction_id atau topup_id
    ├── balanceBefore: Money
    ├── balanceAfter: Money
    └── createdAt: timestamp
```

### Identity Context

```
User (Aggregate Root)
├── userId: UserID             [Value Object]
├── email: Email               [Value Object] — format validated
├── role: UserRole             [Value Object] USER | VERIFIED_MERCHANT | VENDOR | EO | ADMIN
│
├── Profile                    [Entity]
│   ├── name
│   ├── phone
│   └── loyaltyTier: Tier     BRONZE | SILVER | GOLD | PLATINUM
│
└── KYCDocument                [Entity]
    ├── idCardUrl
    ├── selfieUrl
    ├── status: KYCStatus      UNVERIFIED | PENDING | VERIFIED | REJECTED
    └── score: float           0.0 - 1.0 dari backend-ai
```

### Vendor Context

```
Vendor (Aggregate Root)
├── vendorId: VendorID         [Value Object]
├── userId: UserID             [Value Object]
├── businessName: string
├── status: VendorStatus       PENDING | ACTIVE | SUSPENDED
│
├── Category[]                 [Entity]
│   └── categoryId → taxonomy
│
└── Service[]                  [Entity]
    ├── serviceName
    ├── priceFrom: Money
    └── description
```

---

## Domain Events

Semua event menggunakan format: **NounPastTense** (Bahasa Inggris, past tense)

### Identity Events
```
UserRegistered          { userId, email, timestamp }
UserKYCSubmitted        { userId, idCardUrl, selfieUrl }
UserKYCApproved         { userId, adminId, timestamp }
UserKYCRejected         { userId, adminId, reason }
UserRoleUpgraded        { userId, fromRole, toRole }
UserLoyaltyTierChanged  { userId, fromTier, toTier }
```

### Transaction Events
```
TransactionCreated      { txId, type, buyerId, sellerId, amount }
EscrowLocked            { txId, amount, lockedAt }
EscrowReleased          { txId, amount, recipientId, fee }
EscrowFrozen            { txId, disputeId, frozenAt }
TransactionCancelled    { txId, cancelledBy, reason }
TransactionCompleted    { txId, completedAt }

MilestoneSubmitted      { txId, milestoneId, evidenceUrl }
MilestoneApproved       { txId, milestoneId, approvedBy }
MilestonePayoutReleased { txId, milestoneId, amount }

DisputeOpened           { txId, disputeId, openedBy, reason }
DisputeEvidenceAdded    { disputeId, addedBy, evidenceUrl }
DisputeResolved         { disputeId, adminId, resolution, refundAmount }
```

### Financial Events
```
WalletTopUpInitiated    { walletId, amount, midtransOrderId }
WalletTopUpCompleted    { walletId, amount, ledgerId }
WalletWithdrawRequested { walletId, amount, bankAccount }
WalletWithdrawApproved  { walletId, amount, adminId }
FeeDeducted             { txId, feeAmount, platformWalletId }
```

### Vendor Events
```
VendorRegistered        { vendorId, userId, businessName }
VendorVerified          { vendorId, adminId }
VendorAllocated         { vendorId, eventTxId, contractAmount }
VendorPayoutRequested   { vendorId, eventTxId, amount, invoiceUrl }
VendorPayoutApproved    { vendorId, eoId, amount }
VendorPayoutReleased    { vendorId, amount, ledgerId }
```

### Audit Events
```
AuditLogRecorded        { txId, chainTxHash, blockNumber, network }
```

---

## Context Map — Hubungan Antar Bounded Context

```
Identity ──[UserID]──► Transaction    Transaksi butuh validasi user & KYC
Identity ──[UserID]──► Financial      Wallet terhubung ke user
Transaction ──[triggers]──► Financial  EscrowLocked/Released → mutasi wallet
Transaction ──[triggers]──► AI         FraudScore diminta sebelum EscrowLocked
Transaction ──[triggers]──► Audit      TransactionCompleted → AuditLogRecorded
Transaction ──[triggers]──► Notification Event apapun → notifikasi ke user
Vendor ──[VendorAllocation]──► Transaction  Kontrak vendor masuk ke escrow event
Financial ──[snapshot]──► Audit        Mutasi keuangan penting dicatat on-chain
```

---

## Mapping ke Kode (Clean Architecture)

```
DDD Concept          │ Lokasi di Kode
─────────────────────┼────────────────────────────────────────────
Aggregate Root       │ internal/domain/*.go  (struct utama)
Value Object         │ internal/domain/*.go  (tipe primitif typed)
Domain Event         │ internal/domain/worker.go (event definitions)
Repository Interface │ internal/domain/*.go  (interface contract)
Repository Impl      │ internal/repository/*.go
Use Case / Service   │ internal/usecase/*.go (business logic)
Application Service  │ internal/delivery/handlers/*.go
Bounded Context      │ Setiap folder domain di usecase/ adalah 1 context
```

---

## Aturan Invariant Domain

Invariant adalah kondisi yang TIDAK BOLEH dilanggar — enforce di usecase layer:

```
INV-01: availableBalance tidak boleh negatif
INV-02: lockedBalance tidak boleh dilepas tanpa kondisi yang valid
INV-03: EscrowReleased tidak boleh terjadi jika ada DisputeOpened aktif
INV-04: Milestone hanya bisa Approved setelah MilestoneSubmitted
INV-05: Vendor hanya bisa menerima payout setelah EO approve
INV-06: KYC harus VERIFIED sebelum user bisa jadi VERIFIED_MERCHANT/VENDOR/EO
INV-07: idempotencyKey unik per operasi finansial — duplicate ditolak
INV-08: FraudScore >= FRAUD_UNSAFE_THRESHOLD → EscrowLocked diblokir
```