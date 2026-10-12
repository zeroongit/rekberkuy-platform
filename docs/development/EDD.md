# Event-Driven Design Document (EDD)
# RekberKuy Platform

> Dokumen ini mendeskripsikan sistem dari sudut pandang KEJADIAN (event).
> Baca DDD.md terlebih dahulu untuk memahami domain events yang digunakan.

---

## State Machine — Transaction Status

Semua jenis transaksi (Goods, Services, Events) menggunakan
state machine yang sama sebagai fondasi:

```
                    ┌─────────────────┐
                    │      DRAFT      │ ← dibuat tapi belum bayar
                    └────────┬────────┘
                             │ buyerPays()
                             ▼
                    ┌─────────────────┐
                    │ PENDING_PAYMENT │ ← menunggu konfirmasi payment gateway
                    └────────┬────────┘
                             │ paymentConfirmed()
                             ▼
                    ┌─────────────────┐
          ┌────────►│     ACTIVE      │ ← escrow terkunci, transaksi berjalan
          │         └────────┬────────┘
          │                  │
          │         ┌────────┴────────┐
          │         │                 │
          │         ▼                 ▼
          │  ┌────────────┐   ┌──────────────┐
          │  │IN_PROGRESS │   │   DISPUTED   │ ← sengketa dibuka
          │  └─────┬──────┘   └──────┬───────┘
          │        │                 │
          │        ▼                 ▼
          │  ┌────────────┐   ┌──────────────┐
          │  │ COMPLETED  │   │   RESOLVED   │ ← admin memutuskan
          │  └────────────┘   └──────────────┘
          │
          └──────── CANCELLED ◄── sebelum escrow terkunci
```

---

## Event Flow per Domain

### 1. Transaksi Barang (Goods)

```
Waktu  Actor          Event                        Efek
──────────────────────────────────────────────────────────────────────
T+0    Pembeli        TransactionCreated           Status: DRAFT
T+0    System         FraudScoreRequested          → backend-ai dipanggil
T+0    AI             FraudScoreReturned           score diterima core-service
T+0    System         [if score < threshold]
       System         EscrowLocked                 Status: ACTIVE
                                                   walletBuyer.locked += amount
T+0    System         SellerNotified               email + in-app
T+?    Penjual        ShipmentConfirmed            input nomor resi
T+?    System         BuyerNotifiedOfShipment      email + in-app
T+?    Pembeli        DeliveryConfirmed            [HAPPY PATH]
       System         EscrowReleased               walletSeller.available += netPayout
                                                   platformWallet += fee
                                                   Status: COMPLETED
       System         AuditLogRecorded             → Avalanche (gasless)

       ──── ALTERNATIF: Auto-Release ────
T+3d   auto_release   DeliveryAutoConfirmed        timeout 3 hari
       worker         EscrowReleased               sama seperti di atas

       ──── ALTERNATIF: Dispute ────
T+?    Pembeli/       DisputeOpened                EscrowFrozen
       Penjual                                     Status: DISPUTED
T+?    Admin          DisputeResolved              EscrowReleased (ke pembeli/penjual)
                                                   Status: RESOLVED
```

---

### 2. Transaksi Jasa (Services — Milestone)

```
Waktu  Actor          Event                        Efek
──────────────────────────────────────────────────────────────────────
T+0    Klien          TransactionCreated           milestones[] defined
T+0    Penyedia Jasa  TransactionAccepted          Status: ACTIVE
T+0    Klien          EscrowLocked                 total kontrak terkunci
T+?    Penyedia Jasa  MilestoneSubmitted           upload bukti hasil kerja
T+?    Klien          MilestoneApproved            [atau timeout → auto-approve]
       System         MilestonePayoutReleased      walletProvider += milestone.amount
T+?    [repeat per milestone]
T+end  System         TransactionCompleted         semua milestone done
       System         FeeDeducted                  fee dipotong di release terakhir
       System         AuditLogRecorded             → Avalanche

       ──── ALTERNATIF: Revisi ────
T+?    Klien          RevisionRequested            milestone status → REVISION
T+?    Penyedia Jasa  MilestoneResubmitted         upload ulang
```

---

### 3. Transaksi Event

```
Waktu  Actor          Event                        Efek
──────────────────────────────────────────────────────────────────────
T+0    EO             EventCreated                 event dipublikasikan
T+?    Peserta        TicketPurchased              dana masuk escrow event
                                                   tiket QR digenerate
T+?    EO             VendorAllocated              kontrak vendor dibuat
T+?    Sistem         EscrowLockedForVendor        dana vendor terkunci
T+?    Vendor         PayoutRequestSubmitted       upload invoice
T+?    EO             PayoutRequestApproved        review invoice
       System         VendorPayoutReleased         walletVendor += amount
       System         VendorPayoutAuditLogged      → Avalanche
T+?    EO             EventCompleted               event selesai
       System         TicketRevenueReleased        walletEO += net revenue
       System         PlatformFeeDeducted
       System         AuditLogRecorded             → Avalanche

       ──── ALTERNATIF: Event Dibatalkan ────
T+?    EO             EventCancelled
       System         AllTicketsRefunded           loop: tiap pembeli direfund
       System         AllVendorContractsCancelled  negosiasi pembatalan
```

---

### 4. Wallet & Financial

```
Waktu  Actor          Event                        Efek
──────────────────────────────────────────────────────────────────────
T+0    User           TopUpInitiated               order dibuat di Midtrans
T+?    Midtrans       PaymentWebhookReceived       core-service validasi signature
       System         WalletTopUpCompleted         walletUser.available += amount
                                                   ledger entry: CREDIT
       System         UserNotifiedTopUp            email konfirmasi

T+0    User           WithdrawRequested            form + rekening tujuan
T+?    Admin          WithdrawApproved             manual review (MVP)
       System         WalletWithdrawProcessed      walletUser.available -= amount
                                                   ledger entry: DEBIT
       System         [manual bank transfer]
```

---

### 5. KYC & Identity

```
Waktu  Actor          Event                        Efek
──────────────────────────────────────────────────────────────────────
T+0    User           KYCSubmitted                 upload KTP + selfie → Supabase
T+0    System         KYCPrescreeningRequested     → backend-ai (vision model)
T+0    AI             KYCPrescreeningCompleted     score dikembalikan
T+?    Admin          KYCReviewed                  [approve / reject]
       System         KYCApproved                  user.kycStatus = VERIFIED
                      UserRoleUpgraded             USER → VERIFIED_MERCHANT/VENDOR/EO
       System         UserNotifiedKYC              email hasil KYC

       ──── Jika ditolak ────
       System         KYCRejected                  kycStatus = REJECTED
       System         UserNotifiedKYCRejection     email + alasan
       User           [max 3x pengajuan ulang]
```

---

## Event Handler Map

Siapa yang menangani setiap event di kode:

| Event | Handler | Lokasi |
|-------|---------|--------|
| `TransactionCreated` | `transaction_goods_usecase.go` | Sync |
| `FraudScoreRequested` | `fraud/client.go` → backend-ai | Sync |
| `EscrowLocked` | `wallet_usecase.go` + `transaction_repository.go` | Sync (DB tx) |
| `EscrowReleased` | `wallet_usecase.go` + `transaction_repository.go` | Sync (DB tx) |
| `DeliveryAutoConfirmed` | `auto_release_worker.go` | Async (background job) |
| `MilestoneApproved` | `transaction_services_usecase.go` | Sync |
| `AuditLogRecorded` | `relayer/relayer.go` → Avalanche | Async (fire-and-forget) |
| `PaymentWebhookReceived` | `wallet_handler.go` (Midtrans webhook) | Async |
| `KYCPrescreeningRequested` | `kyc_usecase.go` → backend-ai | Sync |
| `UserLoyaltyTierChanged` | `crm_worker.go` | Async (monthly cron) |

---

## Worker Schedule

| Worker | Trigger | Event yang Diproses |
|--------|---------|---------------------|
| `auto_release_worker` | Setiap 1 jam | `DeliveryAutoConfirmed` untuk tx > 3 hari |
| `crm_worker` | Setiap 1 bulan (hari 1) | `UserLoyaltyTierChanged` |

---

## Aturan Event

```
RULE-E01: Event tidak boleh diubah setelah dicatat (immutable)
RULE-E02: EscrowReleased tidak boleh terjadi jika status = DISPUTED
RULE-E03: AuditLogRecorded bersifat fire-and-forget — kegagalan blockchain
          TIDAK boleh rollback transaksi utama (retry via queue)
RULE-E04: PaymentWebhookReceived wajib validasi signature Midtrans sebelum diproses
RULE-E05: Semua operasi wallet wajib dalam database transaction (ACID)
RULE-E06: idempotencyKey dicek sebelum setiap event finansial diproses
```