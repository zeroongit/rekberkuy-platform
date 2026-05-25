# CLAUDE.md — RekberKuy Platform

Panduan ini membantu Claude AI memahami konteks, arsitektur, dan konvensi kode project RekberKuy. Baca seluruh file ini sebelum membuat perubahan apapun.

---

## 🧠 Ringkasan Project

**RekberKuy** adalah platform rekening bersama (escrow) untuk transaksi **Barang**, **Jasa**, dan **Event**. Setiap transaksi dicatat sebagai audit log immutable di blockchain Avalanche. Platform ini juga menyediakan marketplace vendor untuk Event Organizer (EO).

**Tiga domain bisnis utama:**
- Jual beli barang (fisik & digital)
- Jasa (freelance / profesional)
- Pengadaan event + vendor marketplace

---

## 📁 Struktur Project

```
rekberkuy-platform/
├── apps/
│   ├── core-service/     # Backend Go — clean architecture
│   └── dashboard-web/     # Frontend Next.js v16
├── blockchain/           # Smart contract audit log — Hardhat v3
├── e2e-qa/               # QA Automation & E2E testing
├── docs/
│   └── usecase.md        # Use case & alur bisnis — BACA INI DULU
├── .gitlab-ci.yml        # CI/CD pipeline
└── CLAUDE.md             # ← kamu sedang membaca ini
```

---

## ⚙️ Perintah Penting

### Backend (Go)
```bash
cd apps/core-backend
go mod tidy                        # Install dependencies
go run ./cmd/server/main.go        # Jalankan server
go test ./...                      # Jalankan semua test
go build -o bin/server ./cmd/server/main.go  # Build binary
```

### Frontend (Next.js)
```bash
cd apps/web-frontend
npm install                        # Install dependencies
npm run dev                        # Dev server (localhost:3000)
npm run build                      # Production build
npm run lint                       # Lint check
npm run test                       # Unit test (Jest)
```

### Blockchain (Hardhat)
```bash
cd blockchain
npm install
npx hardhat compile                # Kompilasi smart contract
npx hardhat test                   # Jalankan test kontrak
npx hardhat node                   # Jalankan local blockchain (port 8545)
npx hardhat ignition deploy ./ignition/modules/Counter.ts --network localhost
```

### QA / E2E
```bash
cd e2e-qa
npm install
npm run test                       # Jalankan semua skenario QA
```

---

## 🏛️ Arsitektur Backend (Clean Architecture)

Layer berurutan — **jangan skip layer, jangan import terbalik**:

```
delivery/http/  →  usecase/  →  repository/  →  domain/
```

| Layer | Lokasi | Tanggung Jawab |
|-------|--------|----------------|
| `domain/` | `internal/domain/` | Struct entity & interface contract |
| `repository/` | `internal/repository/` | Akses database — implementasi interface domain |
| `usecase/` | `internal/usecase/` | Business logic murni — tidak boleh tahu soal HTTP |
| `delivery/` | `internal/delivery/http/` | Handler HTTP — hanya parsing request & response |

### Domain yang sudah ada
- `user.go` — Entity pengguna & struct CRM Loyalty Tiering (target GMV/streak)
- `transaction.go` — Entity transaksi escrow utama & state status
- `finance.go` — Entity penampung hasil kalkulasi audit, fee platform, bonus, dan auto-refund
- `vendor.go` — Entity profil vendor mitra & skema alokasi anggaran event
- `wallet.go` — Entity dompet, mutasi saldo RekberPay, & log transaksi logistik

### Aturan dependency
- `domain/` tidak boleh import package lain dalam project
- `usecase/` boleh import `domain/` saja
- `repository/` boleh import `domain/` saja
- `delivery/` boleh import `usecase/` dan `domain/`

---

## 🌐 Arsitektur Frontend (Next.js v16)

- Gunakan **App Router** (`src/app/`) — bukan Pages Router
- Komponen UI dari **Shadcn/UI** — jangan buat komponen UI dari scratch jika sudah ada di Shadcn
- Styling hanya dengan **Tailwind CSS** — tidak pakai CSS module atau styled-components
- Semua komponen harus **TypeScript** — tidak ada file `.js` atau `.jsx`
- Gunakan **Server Components** by default; tambahkan `"use client"` hanya jika benar-benar butuh interaktivitas

---

## ⛓️ Blockchain — Peran & Batasan

> **PENTING:** Smart contract di folder `blockchain/` berfungsi **HANYA sebagai audit log transaksi**. Bukan untuk menyimpan dana, bukan escrow on-chain, bukan logika bisnis.

**Yang boleh dilakukan smart contract:**
- Mencatat hash/ID transaksi yang sudah selesai
- Menyimpan timestamp & status akhir transaksi
- Emit event untuk keperluan indexing & transparansi

**Yang TIDAK boleh ada di smart contract:**
- Logika escrow atau penahan dana
- Logika bisnis apapun (kalkulasi fee, validasi, dll)
- Data sensitif pengguna

**Network target:** Avalanche C-Chain (Fuji Testnet untuk development, Mainnet untuk production)

---

## 🧪 Testing

| Jenis Test | Tool | Lokasi |
|------------|------|--------|
| Unit test frontend | Jest | `apps/web-frontend/__tests__/` |
| E2E & QA Automation | Playwright | `e2e-qa/` |
| API test | Postman | `docs/api/` (collection) |
| Smart contract test | Hardhat | `blockchain/test/` |
| Backend unit test | Go test | `apps/core-backend/**/*_test.go` |

**Aturan testing:**
- Setiap usecase baru **wajib** punya unit test
- Setiap endpoint baru **wajib** masuk koleksi Postman
- Alur transaksi utama (barang/jasa/event) **wajib** punya skenario E2E di `e2e-qa/`

---

## 📐 Konvensi Kode

### Go (Backend)
- Nama file: `snake_case` (contoh: `transaction_usecase.go`)
- Nama struct & interface: `PascalCase`
- Nama fungsi ekspor: `PascalCase`, fungsi internal: `camelCase`
- Error handling: selalu return `error`, jangan `panic` kecuali di `main.go`
- Interface didefinisikan di `domain/`, diimplementasikan di `repository/` atau `usecase/`
- Gunakan `context.Context` sebagai parameter pertama di semua fungsi yang menyentuh I/O

```go
// ✅ Benar
func (u *transactionUsecase) CreateTransaction(ctx context.Context, req domain.Transaction) (domain.Transaction, error) {}

// ❌ Salah — tidak ada context, tidak return error
func CreateTx(req domain.Transaction) domain.Transaction {}
```

### TypeScript (Frontend)
- Nama file komponen: `PascalCase.tsx` (contoh: `TransactionCard.tsx`)
- Nama file utility/hook: `camelCase.ts` (contoh: `useTransaction.ts`)
- Selalu definisikan tipe — hindari `any`
- Gunakan `interface` untuk props komponen, `type` untuk union/intersection
- Nama komponen harus deskriptif dan mencerminkan domain bisnis

```tsx
// ✅ Benar
interface TransactionCardProps {
  transactionId: string
  status: 'pending' | 'completed' | 'disputed'
}

// ❌ Salah
const Card = ({ data }: { data: any }) => {}
```

### Solidity (Smart Contract)
- Fungsi hanya untuk **write** (catat transaksi) dan **read** (baca log)
- Emit event setiap ada pencatatan baru
- Jangan gunakan `mapping` yang menyimpan data sensitif

---

## 🔑 Environment Variables

Jangan pernah hardcode secrets. Semua config ada di `apps/core-backend/.env`.

Variabel krusial yang harus ada:
- `DATABASE_URL` — koneksi PostgreSQL via Supabase
- `SUPABASE_SERVICE_ROLE_KEY` — jangan expose ke frontend
- `DEPLOYER_PRIVATE_KEY` — private key wallet deployment blockchain, **sangat sensitif**
- `AVALANCHE_RPC_URL` — endpoint RPC Avalanche

---

## 🚫 Hal yang Tidak Boleh Dilakukan

- ❌ Jangan commit file `.env` ke repository
- ❌ Jangan hardcode URL, port, atau credentials di dalam kode
- ❌ Jangan tambahkan logika bisnis di layer `delivery/http/`
- ❌ Jangan import `usecase/` dari `domain/` (melanggar clean architecture)
- ❌ Jangan gunakan `any` di TypeScript kecuali benar-benar tidak ada pilihan lain
- ❌ Jangan tambahkan logika escrow atau penyimpanan dana ke smart contract
- ❌ Jangan push langsung ke branch `main` atau `develop` — selalu via Merge Request

---

## ✅ Checklist Sebelum Membuat Perubahan

Sebelum menulis kode, pastikan kamu sudah:
- [ ] Membaca `docs/usecase.md` untuk memahami alur bisnis yang relevan
- [ ] Memahami layer mana yang perlu diubah (domain / repository / usecase / delivery)
- [ ] Tidak melanggar aturan dependency antar layer
- [ ] Menyiapkan test untuk kode baru
- [ ] Menggunakan nama yang konsisten dengan domain bisnis yang sudah ada

---

## 📚 Referensi Penting

- Alur bisnis & use case: [`docs/usecase.md`](./docs/usecase.md)
- Domain model: [`apps/core-backend/internal/domain/`](./apps/core-backend/internal/domain/)
- Panduan frontend: [`apps/web-frontend/CLAUDE.md`](./apps/web-frontend/CLAUDE.md)
- CI/CD pipeline: [`.gitlab-ci.yml`](./.gitlab-ci.yml)