# 🔐 RekberKuy Platform

> Platform Rekening Bersama (Escrow) modern untuk transaksi **Barang**, **Jasa**, dan **Event** yang aman, transparan, dan terpercaya.

[![Next.js](https://img.shields.io/badge/Next.js-v16-black?style=flat-square&logo=next.js)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Avalanche](https://img.shields.io/badge/Blockchain-Avalanche-E84142?style=flat-square&logo=avalanche)](https://www.avax.network/)
[![Supabase](https://img.shields.io/badge/Supabase-PostgreSQL-3ECF8E?style=flat-square&logo=supabase)](https://supabase.com/)
[![License](https://img.shields.io/badge/License-Proprietary-red?style=flat-square)]()

---

## 📖 Tentang RekberKuy

**RekberKuy** adalah platform escrow all-in-one yang menghubungkan pembeli dan penjual dalam ekosistem transaksi yang aman. Platform ini menjalankan tiga domain utama secara bersamaan:

| Domain | Deskripsi |
|--------|-----------|
| 🛍️ **Barang** | Transaksi jual beli produk fisik maupun digital dengan jaminan escrow |
| 🛠️ **Jasa** | Perantara pembayaran antara klien dan penyedia jasa profesional |
| 🎪 **Event** | Sistem pengadaan event lengkap dengan marketplace vendor terintegrasi |

Dana pembeli akan ditahan oleh sistem dan hanya dilepaskan ke penjual/penyedia layanan setelah transaksi dikonfirmasi selesai oleh semua pihak. Setiap transaksi dicatat secara permanen di **blockchain Avalanche** melalui smart contract yang berfungsi murni sebagai **audit log** — memastikan seluruh riwayat transaksi transparan, tidak dapat dimanipulasi, dan dapat diverifikasi oleh siapa pun.

---

## 🏗️ Struktur Folder

```
rekberkuy-platform/
│
├── apps/
│   ├── core-service/                     # Backend Utama - Golang 1.25
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go               # Entry point server
│   │   ├── config/                       # Konfigurasi aplikasi
│   │   ├── internal/
│   │   │   ├── delivery/
│   │   │   │   └── http/                 # HTTP handler & routing
│   │   │   ├── domain/                   # Domain model & entity (Murni IDR, Bebas Crypto Address)
│   │   │   │   ├── transaction.go        # Domain transaksi escrow & state status
│   │   │   │   ├── user.go               # Domain pengguna & CRM Loyalty Tiering
│   │   │   │   ├── vendor.go             # Domain model bisnis sub-vendor & alokasi event
│   │   │   │   └── wallet.go             # Domain dompet, saldo RekberPay, & log mutasi
│   │   │   ├── repository/               # Akses data (database layer)
│   │   │   │   ├── vendor_repository.go
│   │   │   │   └── wallet_repository.go
│   │   │   └── usecase/                  # Business logic
│   │   │       ├── finance_calculator.go # Kalkulasi biaya & komisi
│   │   │       └── transaction_usecase.go# Alur transaksi escrow
│   │   ├── .env                          # Environment variables
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── dashboard-web/                    # Frontend - Next.js v16
│       ├── src/
│       │   └── app/
│       │       ├── layout.tsx            # Root layout
│       │       ├── page.tsx              # Halaman utama
│       │       ├── globals.css           # Global styles
│       │       └── favicon.ico
│       ├── public/                       # Aset statis (SVG, gambar)
│       ├── AGENTS.md                     # Panduan AI agent
│       ├── CLAUDE.md                     # Panduan Claude AI
│       ├── next.config.ts
│       ├── tailwind.config (postcss.config.mjs)
│       ├── tsconfig.json
│       └── package.json
│
├── blockchain/                           # Audit Transaksi On-Chain - Hardhat v3
│   ├── contracts/
│   │   ├── Counter.sol                   # Smart contract pencatatan transaksi
│   │   └── Counter.t.sol                 # Test contract
│   ├── ignition/
│   │   └── modules/
│   │       └── Counter.ts                # Modul deployment (Hardhat Ignition)
│   ├── scripts/
│   │   └── send-op-tx.ts                 # Script kirim transaksi on-chain
│   ├── test/
│   │   └── Counter.ts                    # Unit test smart contract
│   ├── hardhat.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── e2e-qa/                               # QA Automation & End-to-End Testing
│
├── docs/
│   └── usecase.md                        # Dokumentasi use case & alur bisnis
│
├── .gitlab-ci.yml                        # CI/CD pipeline GitLab
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
| [Go](https://golang.org/) | 1.25 | Backend monolith dengan arsitektur clean architecture |
| [Redis](https://redis.io/) | Latest | Caching, session, & pub/sub messaging |

### Database
| Teknologi | Kegunaan |
|-----------|----------|
| [PostgreSQL](https://www.postgresql.org/) | Database relasional utama |
| [Supabase](https://supabase.com/) | Backend-as-a-Service (auth, realtime, storage) |

### Blockchain
| Teknologi | Versi | Kegunaan |
|-----------|-------|----------|
| [Avalanche](https://www.avax.network/) | - | Blockchain untuk pencatatan audit transaksi secara transparan & immutable |
| [Hardhat](https://hardhat.org/) | v3 | Framework pengembangan & testing smart contract |
| [Solidity](https://soliditylang.org/) | - | Bahasa pemrograman smart contract |

### Testing & QA
| Teknologi | Scope |
|-----------|-------|
| [Jest](https://jestjs.io/) | Unit test frontend |
| [Playwright](https://playwright.dev/) | End-to-end test (E2E) |
| [Postman](https://www.postman.com/) | API testing backend |
| QA Automation | Otomasi skenario pengujian fungsional & regresi (`e2e-qa/`) |

### DevOps
| Teknologi | Kegunaan |
|-----------|----------|
| [GitLab CI/CD](https://docs.gitlab.com/ee/ci/) | Pipeline otomatis build, test, dan deploy |
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
- Autentikasi multi-faktor (MFA)
- Verifikasi identitas (KYC) untuk penjual & EO
- Notifikasi real-time (in-app, email, push notification)
- Manajemen wallet & saldo pengguna
- Kalkulasi biaya & komisi otomatis (`finance_calculator`)
- Dashboard transaksi & riwayat lengkap
- Admin panel untuk moderasi & manajemen platform
- Smart contract audit log di Avalanche — setiap transaksi tercatat permanen & transparan di blockchain

---

## 🚀 Cara Menjalankan Lokal

### Prasyarat

Pastikan sudah terinstall:
- [Node.js](https://nodejs.org/) `>= 20.x`
- [Go](https://golang.org/) `>= 1.25`
- [Docker](https://www.docker.com/) & Docker Compose
- [Git](https://git-scm.com/)

### 1. Clone Repository

```bash
git clone https://gitlab.com/pelajarsantuy1/rekberkuy-platform.git
cd rekberkuy-platform
```

### 2. Setup Environment Variables

```bash
cp apps/core-service/.env.example apps/core-service/.env
# Edit file .env sesuai konfigurasi lokal kamu
```

### 3. Jalankan Backend (Go)

```bash
cd apps/core-backend
go mod tidy
go run ./cmd/server/main.go
```

### 4. Jalankan Frontend (Next.js)

```bash
cd apps/web-frontend
npm install
npm run dev
```

### 5. Jalankan Blockchain (Hardhat)

```bash
cd blockchain
npm install
npx hardhat compile
npx hardhat node        # Jalankan local blockchain node
```

Aplikasi akan berjalan di:
- **Frontend:** `http://localhost:3000`
- **Backend API:** `http://localhost:8080`
- **Hardhat Node:** `http://localhost:8545`

---

## 🧪 Testing

```bash
# Unit test frontend (Jest)
cd apps/web-frontend
npm run test

# E2E & QA Automation
cd e2e-qa
npm run test

# Test smart contract (audit log)
cd blockchain
npx hardhat test

# Test backend API
# Import koleksi Postman dari /docs/ dan jalankan via Postman
```

---

## 📡 Environment Variables

Salin dan isi file `apps/core-backend/.env`:

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

# Blockchain (Avalanche)
AVALANCHE_RPC_URL=
DEPLOYER_PRIVATE_KEY=
CONTRACT_ADDRESS=

# Notifikasi
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=
```

---

## 📋 To-Do List

### 🔴 High Priority
- [ ] Deploy smart contract `Counter` / `TransactionLogger` ke Avalanche Fuji Testnet
- [ ] Implementasi HTTP handler di `internal/delivery/http` untuk semua domain
- [ ] Ekspansi repository khusus untuk manajemen detail sengketa (`dispute_repository`)
- [ ] Integrasi payment gateway (Midtrans) untuk automasi top-up saldo Rupiah
- [ ] Worker Engine: Membuat sistem otomatisasi cron job bulanan (Eksekusi setiap tanggal 1 jam 00:00 untuk evaluasi kasta CRMLoyalty)
- [ ] Rate limiting & Idempotency Key middleware di delivery layer untuk mencegah double-spending

### 🟡 Medium Priority
- [ ] **Vendor Marketplace** — halaman daftar vendor rekanan RekberKuy untuk EO
- [ ] **Vendor Onboarding** — alur pendaftaran vendor ke dalam platform
- [ ] Sistem review & rating untuk penjual, penyedia jasa, dan vendor
- [ ] Notifikasi real-time menggunakan Supabase Realtime
- [ ] Dashboard analitik & laporan transaksi untuk admin
- [ ] Fitur milestone payment untuk modul Jasa
- [ ] Integrasi pengiriman (JNE, J&T, SiCepat) untuk modul Barang
- [ ] Perlengkapan halaman frontend (barang, jasa, event, vendor, dashboard)
- [ ] Pengisian `e2e-qa/` dengan skenario test otomasi lengkap
- [ ] Multi-bahasa (i18n) — Indonesia & Inggris

### 🟢 Low Priority / Enhancement
- [ ] Mobile app (React Native / Flutter)
- [ ] Fitur chat/negosiasi langsung antar pengguna
- [ ] Affiliate & referral program
- [ ] Integrasi e-wallet (OVO, GoPay, Dana)
- [ ] Dark mode UI
- [ ] PWA (Progressive Web App) support
- [ ] Laporan pajak otomatis untuk transaksi

### ✅ Selesai
- [x] Inisialisasi project Next.js v16 dengan Shadcn/UI & Tailwind v4
- [x] Setup backend Go dengan Clean Architecture
- [x] Domain model murni IDR: `user` (CRM Loyalty), `transaction`, `vendor` (Dynamic Allocation), `wallet`
- [x] Usecase: `finance_calculator` (Kalkulator bonus EO & komisi melandai), `transaction_usecase`
- [x] Repository: `vendor_repository` (GetOrCreateCategory dinamis), `wallet_repository` (Serializable Isolation + SELECT FOR UPDATE anti race-condition)
- [x] Setup Hardhat v3 dengan Hardhat Ignition untuk deployment
- [x] Smart contract dasar (`Counter.sol`) & unit test
- [x] Struktur folder `e2e-qa/` untuk QA Automation
- [x] Konfigurasi GitLab CI/CD pipeline (`.gitlab-ci.yml`)
- [x] Dokumentasi use case (`docs/usecase.md`)

---

## 🏛️ Arsitektur Sistem

```
┌──────────────────────────────────────────────────────────┐
│                  Client (Next.js v16)                    │
│               apps/web-frontend/src/app/                 │
└───────────────────────────┬──────────────────────────────┘
                            │ HTTPS / REST API
┌───────────────────────────▼──────────────────────────────┐
│             Core Backend - Go 1.25                       │
│           apps/core-backend/                             │
│                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │  delivery/  │  │   usecase/   │  │  repository/   │  │
│  │    http/    │→ │  (business   │→ │  (data access) │  │
│  │ (handlers)  │  │   logic)     │  │                │  │
│  └─────────────┘  └──────────────┘  └────────────────┘  │
│                          │                   │           │
│                   ┌──────▼──────┐            │           │
│                   │  domain/    │            │           │
│                   │ user        │            │           │
│                   │ transaction │            │           │
│                   │ finance     │            │           │
│                   │ vendor      │            │           │
│                   │ wallet      │            │           │
│                   └─────────────┘            │           │
└──────────────────────────────────────────────┼───────────┘
                                               │
               ┌───────────────────────────────┤
               │                               │
               ▼                               ▼
     ┌──────────────────┐            ┌──────────────────┐
     │   PostgreSQL     │            │      Redis       │
     │  (via Supabase)  │            │    (Cache)       │
     └──────────────────┘            └──────────────────┘
               │
   (catat log transaksi)
               ▼
     ┌──────────────────────┐
     │   Avalanche Network  │
     │  blockchain/         │
     │  (Audit Log Only —   │
     │  immutable on-chain) │
     └──────────────────────┘
```

### Alur Clean Architecture Backend

```
HTTP Request
    │
    ▼
delivery/http/     ← menerima & memvalidasi request
    │
    ▼
usecase/           ← menjalankan business logic
    │
    ▼
repository/        ← baca/tulis ke database
    │
    ▼
domain/            ← definisi entity & interface
```

---

## 🤝 Kontribusi

Project ini bersifat **private** dan hanya untuk tim internal RekberKuy. Untuk berkontribusi:

1. Buat branch baru dari `develop`: `git checkout -b feature/nama-fitur`
2. Commit perubahan: `git commit -m "feat: deskripsi perubahan"`
3. Push ke branch: `git push origin feature/nama-fitur`
4. Buat Merge Request ke branch `develop`

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

---

## 📄 Lisensi

Copyright © 2025 RekberKuy. All rights reserved. — **Proprietary & Confidential.**

---

<div align="center">
  <p>Dibuat dengan ❤️ oleh Tim RekberKuy</p>
  <p><strong>Transaksi Aman, Bisnis Lancar.</strong></p>
</div>