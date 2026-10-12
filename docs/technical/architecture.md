# System Architecture & Data Flows — RekberKuy Platform

## 1. High-Level Architecture
RekberKuy dibangun menggunakan arsitektur modular yang memisahkan antara presentation layer (Next.js), core business logic (Go Clean Architecture), AI inference service (Python FastAPI), dan smart contract blockchain (Avalanche C-Chain).

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (Next.js v16)                   │
│                    Dashboard Web & UI                       │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTPS / JSON
┌──────────────────────────────▼──────────────────────────────┐
│                    Core Service (Go 1.25)                   │
│   Clean Architecture: Handlers → Usecases → Repositories    │
│   Modules: Auth, Wallet, Goods, Services, Events, Dispute   │
└───────┬──────────────────────┬──────────────────────────────┘
        │                      │
        ▼                      ▼
┌──────────────┐       ┌─────────────────────────────────────┐
│  backend-ai  │       │             Blockchain              │
│  (Python)    │       │  Avalanche C-Chain / Fuji Testnet   │
│  Groq LLM    │       │  TransactionLogger.sol (Gasless)    │
└──────────────┘       └─────────────────────────────────────┘
        │                      │
        └──────────┬───────────┘
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│   PostgreSQL (Supabase) · Redis Cache · Midtrans Payment    │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Backend Clean Architecture Layers (`apps/core-service/`)
Layer dalam backend Go terstruktur secara sekuensial dan ketat:
1. **`domain/`**: Entity structs, business rules, dan interface contracts (tidak boleh mengimpor package luar).
2. **`repository/`**: Akses database PostgreSQL menggunakan GORM/sqlx dan implementasi domain interface.
3. **`usecase/`**: Logika bisnis murni (transaksi, wallet, kalkulator fee, dispute), tidak tahu menahu tentang HTTP.
4. **`delivery/handlers/`**: HTTP handlers, routing, dan pemetaan request/response JSON.

---

## 3. Transaction Data Flows

### A. Alur Transaksi Barang & Gasless Audit Log
1. **Inisiasi:** Pembeli membuat transaksi barang (`POST /api/v1/transactions/goods`).
2. **Fraud Check:** Core service memanggil `backend-ai` (`POST /api/v1/fraud/score`) untuk menilai risiko transaksi menggunakan Groq LLM.
3. **Fund Locking:** Jika aman, saldo pembeli didebet dan dana dikunci di escrow (`FUNDS_LOCKED`) menggunakan database transaction ACID (`SELECT FOR UPDATE`).
4. **Fulfillment:** Penjual mengirim barang dan memasukkan nomor resi.
5. **Release & Audit:** Pembeli konfirmasi terima (atau auto-release worker terpicu), dana dicairkan ke penjual, dan asynchronous relayer mencatat hash transaksi ke Avalanche (`TransactionLogger.sol`).

### B. Alur AI-Assisted KYC & Scoring
- `backend-ai` bertindak sebagai **scoring-only service**. Saat user mengajukan KYC (`POST /api/v1/kyc/submit`), core service meminta skor verifikasi ke `backend-ai` dan menyimpannya sebagai referensi (`ai_score`).
- Keputusan mutlak (*verdict*) berada di tangan Admin melalui review manual (`POST /api/v1/admin/kyc/:id/review`).
