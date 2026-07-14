# 🔐 RekberKuy Platform

> Platform Rekening Bersama (Escrow) modern untuk transaksi **Barang**, **Jasa**, dan **Event** yang aman, transparan, dan terpercaya.

[![Next.js](https://img.shields.io/badge/Next.js-v16-black?style=flat-square&logo=next.js)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Python](https://img.shields.io/badge/Python-3.11-3776AB?style=flat-square&logo=python)](https://www.python.org/)
[![Avalanche](https://img.shields.io/badge/Blockchain-Avalanche-E84142?style=flat-square&logo=avalanche)](https://www.avax.network/)
[![Supabase](https://img.shields.io/badge/Supabase-PostgreSQL-3ECF8E?style=flat-square&logo=supabase)](https://supabase.com/)
[![CI](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-2088FF?style=flat-square&logo=github-actions)](https://github.com/features/actions)
[![Test Required](https://img.shields.io/badge/Unit%20Test-Wajib%20di%20Setiap%20PR-orange?style=flat-square&logo=jest)]()
[![License](https://img.shields.io/badge/License-Proprietary-red?style=flat-square)]()

---

> [!WARNING]
> **Setiap perubahan kode wajib disertai unit test.**
> Pull Request tanpa test tidak akan di-review dan tidak akan di-merge ke branch `develop` maupun `main`.
> Lihat panduan lengkap di bagian [🧪 Testing](#-testing) dan [🤝 Kontribusi](#-kontribusi).

---

**RekberKuy** adalah platform escrow all-in-one yang menghubungkan pembeli dan penjual dalam ekosistem transaksi yang aman. Platform ini menjalankan tiga domain utama secara bersamaan:

| Domain | Deskripsi |
|--------|-----------|
| 🛍️ **Barang** | Transaksi jual beli produk fisik maupun digital dengan jaminan escrow |
| 🛠️ **Jasa** | Perantara pembayaran antara klien dan penyedia jasa profesional |
| 🎪 **Event** | Sistem pengadaan event lengkap dengan marketplace vendor terintegrasi |

Dana pembeli ditahan oleh sistem dan hanya dilepaskan ke penjual/penyedia layanan setelah transaksi dikonfirmasi selesai. Setiap transaksi dicatat secara permanen di **blockchain Avalanche** melalui smart contract yang berfungsi murni sebagai **audit log**. Karena regulasi di Indonesia belum melegalkan pembayaran kripto, sistem ini menerapkan **Gasless Transaction** — seluruh proses on-chain berjalan di latar belakang, di mana backend bertindak sebagai *relayer* yang mensubsidi *gas fee*, sehingga pengguna tidak perlu memiliki wallet crypto.

---

## 🏗️ Struktur Folder

```
rekberkuy-platform/
│
├── apps/
│   ├── core-service/                         # Backend Utama — Go 1.25 (Clean Architecture)
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go                   # Entry point & dependency wiring
│   │   ├── config/
│   │   │   └── config.go                     # Konfigurasi aplikasi & environment
│   │   ├── internal/
│   │   │   ├── delivery/
│   │   │   │   └── handlers/
│   │   │   │       ├── auth_middleware.go     # Validasi JWT di setiap request
│   │   │   │       ├── cors_middleware.go     # Konfigurasi CORS
│   │   │   │       ├── idempotency_middleware.go # Pencegahan double-spending
│   │   │   │       ├── kyc_handler.go        # Endpoint KYC & verifikasi identitas
│   │   │   │       ├── transaction_goods_handler.go   # Endpoint transaksi Barang
│   │   │   │       ├── transaction_services_handler.go # Endpoint transaksi Jasa
│   │   │   │       ├── transaction_events_handler.go  # Endpoint transaksi Event
│   │   │   │       ├── user_handlers.go      # Endpoint register, login, profil
│   │   │   │       ├── vendor_handler.go     # Endpoint manajemen vendor
│   │   │   │       └── wallet_handler.go     # Endpoint wallet & top-up
│   │   │   ├── domain/                       # Entity & interface — bebas dependency
│   │   │   │   ├── auth.go                   # Entity autentikasi & session
│   │   │   │   ├── category.go               # Entity kategori transaksi & vendor
│   │   │   │   ├── dispute.go                # Entity sengketa & mediasi
│   │   │   │   ├── finance.go                # Entity keuangan, fee & komisi
│   │   │   │   ├── profile.go                # Entity profil & CRM Loyalty Tiering
│   │   │   │   ├── transaction.go            # Entity transaksi escrow & state machine
│   │   │   │   ├── wallet.go                 # Entity dompet, saldo RekberPay & log mutasi
│   │   │   │   └── worker.go                 # Entity background worker & job scheduling
│   │   │   ├── fraud/
│   │   │   │   └── client.go                 # Interface fraud scoring (→ backend-ai)
│   │   │   ├── relayer/
│   │   │   │   └── relayer.go                # Interface gasless tx (→ Avalanche)
│   │   │   ├── repository/
│   │   │   │   ├── finance_repository.go     # Persistensi kalkulasi fee
│   │   │   │   ├── idempotency_repository.go # Penyimpanan idempotency key
│   │   │   │   ├── kyc_repository.go         # Akses data dokumen & status KYC
│   │   │   │   ├── transaction_repository.go # CRUD transaksi utama
│   │   │   │   ├── transaction_detail_repository.go # Detail item transaksi
│   │   │   │   ├── user_repository.go        # Akses data pengguna
│   │   │   │   ├── vendor_repository.go      # Akses data vendor & kategori dinamis
│   │   │   │   └── wallet_repository.go      # Wallet (Serializable Isolation + SELECT FOR UPDATE)
│   │   │   ├── usecase/
│   │   │   │   ├── finance_calculator.go     # Kalkulasi biaya, komisi & bonus EO
│   │   │   │   ├── kyc_usecase.go            # Alur verifikasi identitas
│   │   │   │   ├── transaction_goods_usecase.go    # Business logic transaksi Barang
│   │   │   │   ├── transaction_services_usecase.go # Business logic transaksi Jasa (milestone)
│   │   │   │   ├── transaction_events_usecase.go   # Business logic transaksi Event
│   │   │   │   ├── user_usecase.go           # Registrasi, login & manajemen pengguna
│   │   │   │   ├── vendor_usecase.go         # Manajemen vendor & alokasi event
│   │   │   │   └── wallet_usecase.go         # Top-up, tarik dana & mutasi saldo
│   │   │   └── worker/
│   │   │       ├── auto_release_worker.go    # Auto-release escrow setelah timeout
│   │   │       └── crm_worker.go             # Evaluasi CRM Loyalty Tiering bulanan
│   │   ├── .env                              # ⚠️ Jangan di-commit
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── dashboard-web/                        # Frontend — Next.js v16
│       ├── src/
│       │   └── app/
│       │       ├── layout.tsx                # Root layout
│       │       ├── page.tsx                  # Halaman utama
│       │       ├── globals.css
│       │       └── favicon.ico
│       ├── public/                           # Aset statis
│       ├── AGENTS.md                         # Panduan AI agent
│       ├── CLAUDE.md                         # Panduan Claude AI
│       ├── next.config.ts
│       ├── postcss.config.mjs
│       ├── tsconfig.json
│       └── package.json
│
├── backend-ai/                               # AI Service — Python (KYC & Fraud Detection)
│                                             # Dipanggil oleh core-service via internal/fraud/client.go
│
├── blockchain/                               # Audit Log On-Chain — Hardhat v3
│   ├── contracts/
│   │   ├── Counter.sol                       # Smart contract pencatatan transaksi (WIP)
│   │   └── Counter.t.sol                     # Test contract
│   ├── ignition/
│   │   └── modules/
│   │       └── Counter.ts                    # Modul deployment (Hardhat Ignition)
│   ├── scripts/
│   │   └── send-op-tx.ts                     # Script kirim transaksi on-chain
│   ├── test/
│   │   └── Counter.ts                        # Unit test smart contract
│   ├── hardhat.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── e2e-qa/                                   # QA Automation & End-to-End Testing
│
├── docs/
│   ├── concept.md                            # Konsep bisnis & model monetisasi
│   └── usecase.md                            # Use case & alur bisnis lengkap
│
├── .github/
│   └── workflows/
│       └── ci.yml                            # GitHub Actions CI/CD pipeline
├── CLAUDE.md                                 # Panduan Claude AI (root)
├── ROADMAP.md                                # Roadmap pengembangan
├── .gitignore
└── README.md
```

---

## 🛠️ Tech Stack

### Frontend
| Teknologi | Versi | Kegunaan |
|-----------|-------|----------|
| [Next.js](https://nextjs.org/) | v16 | Framework React dengan App Router & SSR |
| [Shadcn/UI](https://ui.shadcn.com/) | Latest | Komponen UI berbasis Radix & Tailwind CSS |
| [Tailwind CSS](https://tailwindcss.com/) | v4 | Utility-first CSS framework |
| [TypeScript](https://www.typescriptlang.org/) | v5 | Type safety di seluruh codebase |

### Backend
| Teknologi | Versi | Kegunaan |
|-----------|-------|----------|
| [Go](https://golang.org/) | 1.25 | Core service — Clean Architecture |
| [Python](https://www.python.org/) | 3.11 | AI service — KYC verification & fraud detection (Groq) |
| [Redis](https://redis.io/) | Latest | Caching, session & idempotency key |

### Database
| Teknologi | Kegunaan |
|-----------|----------|
| [PostgreSQL](https://www.postgresql.org/) | Database relasional utama |
| [Supabase](https://supabase.com/) | BaaS — auth, realtime & storage |

### Blockchain
| Teknologi | Versi | Kegunaan |
|-----------|-------|----------|
| [Avalanche](https://www.avax.network/) | — | Audit log transaksi — immutable & transparan |
| [Hardhat](https://hardhat.org/) | v3 | Framework development & testing smart contract |
| [Solidity](https://soliditylang.org/) | — | Bahasa smart contract |

### Testing & QA
| Teknologi | Scope |
|-----------|-------|
| [Jest](https://jestjs.io/) | Unit test frontend |
| [Playwright](https://playwright.dev/) | End-to-end test (E2E) |
| [Postman](https://www.postman.com/) | API testing backend |
| QA Automation | Skenario pengujian fungsional & regresi (`e2e-qa/`) |

### DevOps
| Teknologi | Kegunaan |
|-----------|----------|
| [GitHub Actions](https://github.com/features/actions) | CI/CD pipeline otomatis (`.github/workflows/ci.yml`) |
| [Docker](https://www.docker.com/) | Containerisasi aplikasi |

---

## ✨ Fitur Utama

### 🛍️ Modul Barang
- Listing produk fisik & digital oleh penjual
- Sistem penawaran dan negosiasi harga
- Escrow otomatis saat pembayaran berhasil
- Konfirmasi penerimaan barang oleh pembeli
- Pelepasan dana ke penjual setelah konfirmasi
- Sistem dispute & mediasi jika terjadi sengketa

### 🛠️ Modul Jasa
- Posting kebutuhan jasa oleh klien
- Penawaran dari penyedia jasa (freelancer/vendor)
- Milestone-based payment (pembayaran bertahap)
- Review & rating setelah pekerjaan selesai
- Proteksi untuk kedua pihak melalui escrow

### 🎪 Modul Event
- Pembuatan & manajemen event oleh Event Organizer (EO)
- Sistem pemesanan & ticketing
- **Marketplace Vendor Terintegrasi:**
  - EO dapat memilih vendor rekanan dari dalam platform
  - EO dapat mendaftarkan vendor eksternal/pilihan sendiri
- Manajemen kontrak vendor dengan escrow pembayaran
- Dashboard monitoring progres persiapan event

### 🔐 Fitur Platform
- Autentikasi JWT & session management
- Verifikasi identitas (KYC) dengan AI scoring via Groq
- Fraud detection otomatis sebelum release escrow
- Idempotency key middleware — pencegahan double-spending
- Notifikasi real-time (in-app, email, push notification)
- Manajemen wallet & saldo (Serializable Isolation + SELECT FOR UPDATE)
- Kalkulasi biaya & komisi otomatis (`finance_calculator`)
- CRM Loyalty Tiering — evaluasi otomatis setiap bulan (`crm_worker`)
- Auto-release escrow setelah timeout (`auto_release_worker`)
- Smart contract audit log di Avalanche — **Gasless Transaction**, pengguna tidak perlu wallet crypto

---

## 🏛️ Arsitektur Sistem

```
┌──────────────────────────────────────────────────────────┐
│                  Client (Next.js v16)                    │
│               apps/dashboard-web/src/app/                │
└───────────────────────────┬──────────────────────────────┘
                            │ HTTPS / REST API
┌───────────────────────────▼──────────────────────────────┐
│                Core Service — Go 1.25                    │
│                  apps/core-service/                      │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │              delivery/handlers/                    │  │
│  │  auth · cors · idempotency · kyc                  │  │
│  │  transaction_goods · transaction_services          │  │
│  │  transaction_events · user · vendor · wallet       │  │
│  └──────────────────────┬─────────────────────────────┘  │
│                         ▼                                │
│  ┌────────────────────────────────────────────────────┐  │
│  │                   usecase/                         │  │
│  │  goods · services · events · kyc · user           │  │
│  │  vendor · wallet · finance_calculator             │  │
│  └──────────────────────┬─────────────────────────────┘  │
│                         ▼                                │
│  ┌────────────────────────────────────────────────────┐  │
│  │                  repository/                       │  │
│  │  transaction · transaction_detail · user · kyc    │  │
│  │  vendor · wallet · finance · idempotency          │  │
│  └──────────────────────┬─────────────────────────────┘  │
│                         ▼                                │
│  ┌────────────────────────────────────────────────────┐  │
│  │                   domain/                          │  │
│  │  auth · profile · transaction · dispute · finance  │  │
│  │  category · wallet · worker                        │  │
│  └─────────────────────────────────────────────────── ┘  │
│                         │                                │
│           ┌─────────────┴─────────────┐                  │
│           ▼                           ▼                  │
│   ┌──────────────┐           ┌──────────────────┐        │
│   │ fraud/       │           │ relayer/         │        │
│   │ client.go    │           │ relayer.go       │        │
│   │ (interface)  │           │ (interface)      │        │
│   └──────┬───────┘           └────────┬─────────┘        │
└──────────┼────────────────────────────┼──────────────────┘
           │                            │
           ▼                            ▼
┌──────────────────┐         ┌──────────────────────┐
│   backend-ai/    │         │   Avalanche Network  │
│   (Python)       │         │   blockchain/        │
│   KYC + Fraud    │         │   (Audit Log Only —  │
│   Detection      │         │   Gasless Relayer)   │
│   via Groq       │         └──────────────────────┘
└──────────────────┘

         ┌─────────────────┐     ┌──────────────┐
         │   PostgreSQL    │     │    Redis     │
         │  (via Supabase) │     │   (Cache &   │
         └─────────────────┘     │  Idempotency)│
                                 └──────────────┘
```

### Alur Clean Architecture

```
HTTP Request
    │
    ▼
delivery/handlers/   ← parsing request, auth middleware, validasi payload
    │
    ▼
usecase/             ← business logic (goods/services/events dipisah)
    │           │
    │           ├──► fraud/client.go   → backend-ai (Python/Groq)
    │           └──► relayer/relayer.go → blockchain Avalanche
    ▼
repository/          ← akses PostgreSQL & Redis
    │
    ▼
domain/              ← entity & interface — tidak boleh import layer lain
```

---

## 🚀 Cara Menjalankan Lokal

### Prasyarat
- [Go](https://golang.org/) `>= 1.25`
- [Node.js](https://nodejs.org/) `>= 20.x`
- [Python](https://www.python.org/) `>= 3.11`
- [Docker](https://www.docker.com/) & Docker Compose
- [Git](https://git-scm.com/)

### 1. Clone Repository

```bash
git clone https://github.com/zeroongit/rekberkuy-platform.git
cd rekberkuy-platform
```

### 2. Setup Environment Variables

```bash
cp apps/core-service/.env.example apps/core-service/.env
# Edit .env sesuai konfigurasi lokal
```

### 3. Jalankan Backend (Go)

```bash
cd apps/core-service
go mod tidy
go run ./cmd/server/main.go
```

### 4. Jalankan Frontend (Next.js)

```bash
cd apps/dashboard-web
npm install
npm run dev
```

### 5. Jalankan AI Service (Python)

```bash
cd backend-ai
pip install -r requirements.txt
uvicorn main:app --reload --port 8081
```

### 6. Jalankan Blockchain Node (Hardhat)

```bash
cd blockchain
npm install
npx hardhat compile
npx hardhat node   # Local blockchain di port 8545
```

Aplikasi berjalan di:
- **Frontend:** `http://localhost:3000`
- **Core Service (Go):** `http://localhost:8080`
- **AI Service (Python):** `http://localhost:8081`
- **Hardhat Node:** `http://localhost:8545`

---

## 🧪 Testing

> [!IMPORTANT]
> **Aturan wajib sebelum membuat Pull Request:**
> - **Backend (Go):** Setiap `usecase` dan `repository` baru wajib punya file `*_test.go`
> - **Frontend (Next.js):** Setiap komponen dan utility baru wajib punya `*.test.tsx` / `*.test.ts`
> - **AI Service (Python):** Setiap endpoint baru wajib punya `test_*.py`
> - **Smart Contract:** Setiap fungsi kontrak baru wajib punya test di `blockchain/test/`
> - **Alur kritis** (transaksi, wallet, escrow): wajib punya skenario E2E di `e2e-qa/`

```bash
# Unit test backend (Go)
cd apps/core-service
go test ./...

# Unit test dengan coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Unit test dengan race detector (wajib untuk kode concurrent)
go test -race ./...

# Unit test frontend (Jest)
cd apps/dashboard-web
npm run test
npm run test -- --coverage

# Unit test AI service (Python)
cd backend-ai
pytest

# E2E & QA Automation
cd e2e-qa
npm run test

# Test smart contract
cd blockchain
npx hardhat test

# API test — import koleksi Postman dari /docs/
```

### Coverage Minimum

| Layer | Minimum |
|-------|---------|
| `usecase/` (Go) | 80% |
| `repository/` (Go) | 70% |
| `delivery/handlers/` (Go) | 60% |
| `lib/utils/` (Frontend) | 90% |
| `backend-ai/` endpoints | 70% |

> [!CAUTION]
> PR yang menurunkan coverage di bawah batas minimum akan **ditolak** meski fungsionalitasnya benar.

---

## 📡 Environment Variables

Salin dan isi `apps/core-service/.env`:

```env
# App
APP_ENV=development
APP_PORT=8080

# Supabase
SUPABASE_URL=
SUPABASE_ANON_KEY=
SUPABASE_SERVICE_ROLE_KEY=

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/rekberkuy

# Redis
REDIS_URL=redis://localhost:6379

# AI Service
AI_SERVICE_URL=http://localhost:8081
GROQ_API_KEY=

# Blockchain (Avalanche)
AVALANCHE_RPC_URL=
DEPLOYER_PRIVATE_KEY=
CONTRACT_ADDRESS=

# Payment Gateway
MIDTRANS_SERVER_KEY=
MIDTRANS_CLIENT_KEY=

# Notifikasi
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=
```

---

## 📋 To-Do List

### 🔴 High Priority
- [ ] **`backend-ai/`** — implementasi FastAPI service untuk KYC scoring & fraud detection via Groq
- [ ] **`blockchain/`** — ganti `Counter.sol` dengan `TransactionLogger.sol` yang sesuai konsep audit log
- [ ] Deploy `TransactionLogger` ke Avalanche Fuji Testnet
- [ ] Implementasi `relayer.go` — wiring nyata ke Avalanche via `go-ethereum`
- [ ] Implementasi `fraud/client.go` — wiring nyata ke `backend-ai`
- [ ] Integrasi Midtrans untuk top-up saldo wallet
- [ ] Unit test semua usecase & repository (`*_test.go`)
- [ ] Seed database untuk testing (`scripts/seed.sql`)

### 🟡 Medium Priority
- [ ] **Vendor Marketplace** — halaman daftar vendor rekanan untuk EO
- [ ] **Vendor Onboarding** — alur pendaftaran vendor ke platform
- [ ] Sistem review & rating (penjual, penyedia jasa, vendor)
- [ ] Notifikasi real-time via Supabase Realtime
- [ ] Dashboard analitik & laporan transaksi (admin)
- [ ] Milestone payment modul Jasa
- [ ] Integrasi pengiriman (JNE, J&T, SiCepat) untuk modul Barang
- [ ] Perlengkapan halaman frontend (barang, jasa, event, vendor, dashboard)
- [ ] Pengisian `e2e-qa/` dengan skenario test lengkap
- [ ] `dispute_repository.go` — manajemen detail sengketa

### 🟢 Low Priority / Enhancement
- [ ] Mobile app (React Native / Flutter)
- [ ] Fitur chat/negosiasi langsung antar pengguna
- [ ] Affiliate & referral program
- [ ] Integrasi e-wallet (OVO, GoPay, Dana)
- [ ] Multi-bahasa (i18n) — Indonesia & Inggris
- [ ] Dark mode UI
- [ ] PWA (Progressive Web App) support
- [ ] Laporan pajak otomatis

### ✅ Selesai
- [x] Setup monorepo: `apps/`, `blockchain/`, `backend-ai/`, `e2e-qa/`, `docs/`
- [x] Inisialisasi Next.js v16 + Shadcn/UI + Tailwind v4 (`dashboard-web`)
- [x] Setup Go backend dengan Clean Architecture (`core-service`)
- [x] Domain model: `auth`, `profile`, `transaction`, `dispute`, `finance`, `category`, `wallet`, `worker`
- [x] Refaktorisasi transaksi — dipisah per segmen: `goods`, `services`, `events` (domain, usecase, handler)
- [x] Usecase: `finance_calculator`, `transaction_goods`, `transaction_services`, `transaction_events`, `user`, `vendor`, `wallet`, `kyc`
- [x] Repository: `transaction`, `transaction_detail`, `user`, `kyc`, `vendor`, `wallet`, `finance`, `idempotency`
- [x] Handler: `auth_middleware`, `cors_middleware`, `idempotency_middleware`, `kyc`, `transaction_goods`, `transaction_services`, `transaction_events`, `user`, `vendor`, `wallet`
- [x] Worker: `auto_release_worker` (timeout escrow), `crm_worker` (evaluasi Loyalty Tiering)
- [x] Stub interface: `internal/fraud/client.go` & `internal/relayer/relayer.go`
- [x] Setup Hardhat v3 + Hardhat Ignition + smart contract dasar
- [x] Struktur `e2e-qa/` & `backend-ai/`
- [x] Migrasi dari GitLab ke GitHub
- [x] GitHub Actions CI/CD pipeline (`.github/workflows/ci.yml`)
- [x] Dokumentasi: README, CLAUDE.md, ROADMAP.md, `docs/usecase.md`, `docs/concept.md`

---

## 🤝 Kontribusi

Project ini bersifat **private** dan hanya untuk tim internal RekberKuy. Untuk berkontribusi:

1. Buat branch baru dari `develop`: `git checkout -b feature/nama-fitur`
2. Tulis kode perubahan
3. **Tulis unit test untuk setiap kode yang ditambahkan atau diubah** ← wajib
4. Pastikan semua test lolos: `go test ./...` (Go) / `npm run test` (Frontend) / `pytest` (Python)
5. Commit perubahan: `git commit -m "feat: deskripsi perubahan"`
6. Push ke branch: `git push origin feature/nama-fitur`
7. Buat Pull Request ke branch `develop`

> [!CAUTION]
> Pull Request yang tidak menyertakan unit test **akan langsung ditolak** tanpa proses review. Pastikan setiap fungsi, usecase, dan komponen baru memiliki test yang memadai sebelum membuka PR.

### Konvensi Commit

Gunakan format [Conventional Commits](https://www.conventionalcommits.org/):

```
feat:     Fitur baru
fix:      Perbaikan bug
docs:     Perubahan dokumentasi
style:    Formatting (tidak mengubah logika)
refactor: Refactoring kode
test:     Menambah atau memperbaiki test
chore:    Update dependency, konfigurasi, dll
```

> Selalu branch dari `develop` (`feature/nama-fitur`).

---

## 📄 Lisensi

Copyright © 2026 RekberKuy. All rights reserved. — **Proprietary & Confidential.**

---