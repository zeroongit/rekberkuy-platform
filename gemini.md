# 🧠 Gemini Code Assist - Global Documentation

Panduan utama (Global Rules) untuk pengembangan **RekberKuy Platform**. Baca file ini untuk memahami gambaran besar sistem sebelum menulis kode. 

**PENTING**: Terdapat juga file `gemini.md` spesifik di masing-masing folder (`apps/core-service/`, `apps/dashboard-web/`, dan `blockchain/`) untuk detail teknis per modul.

---

## 🚀 Ringkasan Project
**RekberKuy** adalah platform rekening bersama (escrow) terpercaya untuk transaksi **Barang**, **Jasa**, dan **Event** (beserta Vendor Marketplace).
- **Prinsip Utama**: Seluruh transaksi dan logika uang (Murni IDR) berjalan di **Backend**.
- **Transparansi**: Setiap transaksi selesai dicatat sebagai *Audit Log* permanen di **Blockchain Avalanche**.
- **Kepatuhan Hukum (Gasless)**: Karena pembayaran kripto belum legal di Indonesia, sistem berjalan murni dengan mata uang Rupiah (IDR). Interaksi dengan blockchain diterapkan melalui **Gasless Transaction** yang dieksekusi oleh Backend sebagai relayer.

---

## 🏗️ Struktur & Arsitektur Sistem Utama

### 1. Backend (`apps/core-service/`)
- **Stack**: Golang 1.25, PostgreSQL (via Supabase), Redis.
- **Arsitektur**: Clean Architecture Murni.
  - **Alur Ketergantungan**: `delivery/http` ➔ `usecase` ➔ `repository` ➔ `domain`.
  - **Aturan Ketat**: Dilarang melompati layer (misal `delivery` memanggil `repository` langsung) atau mengimpor arah terbalik (misal `domain` mengimpor `usecase`).
- **Konteks Keuangan**: Sistem ini murni **Rupiah (IDR)**. Dilarang menggunakan/meminta alamat crypto dari pengguna. Backend secara otomatis mem-broadcast transaksi ke blockchain di belakang layar.

### 2. Frontend (`apps/dashboard-web/`)
- **Stack**: Next.js v16, Tailwind CSS v4, Shadcn/UI, TypeScript.
- **Arsitektur**: Strictly **App Router** (`src/app/`).
- **Aturan Ketat**: 
  - Gunakan *Server Components* secara default.
  - Gunakan komponen Shadcn/UI terlebih dahulu sebelum membuat sendiri.
  - Dilarang menggunakan CSS module atau styled-components (Hanya Tailwind).
  - Hindari penggunaan tipe `any` kecuali terpaksa.

### 3. Blockchain (`blockchain/`)
- **Stack**: Solidity, Hardhat v3, Avalanche C-Chain.
- **Tujuan Khusus**: **HANYA SEBAGAI AUDIT LOG**.
- **Aturan Ketat**: DILARANG meletakkan logika escrow, menahan dana, melakukan kalkulasi biaya, atau menyimpan data PII/sensitif (termasuk di `mapping`) secara on-chain.

---

## 🧪 Konvensi Pengujian (Testing)
- **Backend**: Unit test di Go (`go test`).
- **Frontend**: Unit test dengan Jest (`npm run test`).
- **Smart Contract**: Hardhat Test (`npx hardhat test`).
- **E2E & QA**: Playwright & skenario automasi diletakkan di folder `e2e-qa/`.

Setiap fungsi/usecase baru di backend dan smart contract **WAJIB** memiliki unit test yang menutupi skenario sukses dan gagal.

---

## 🚫 Batasan Mutlak (Do NOT)
1. **JANGAN** hardcode *secrets*, URL database, API keys, atau RPC URLs di dalam kode. Selalu gunakan *Environment Variables* (rujuk `.env.example`).
2. **JANGAN** mengabaikan penanganan `error` di Golang (selalu `return error`, jangan di-`panic`).
3. **JANGAN** pernah mengubah saldo dompet pengguna (`wallet_repository.go`) tanpa menggunakan Database Transaction (`BeginTx`) dan perlindungan *Anti Race Condition* (`SELECT ... FOR UPDATE`).
4. **JANGAN** mengubah file *smart contract* (`Counter.sol` / `TransactionLogger.sol`) menjadi fungsi *hold funds*. Logika pemrosesan pembayaran (Midtrans/RekberPay) mutlak dipegang backend.

---

## 🤝 Git Workflow & Kontribusi
- Gunakan format **Conventional Commits**:
  - `feat:` (Fitur baru)
  - `fix:` (Perbaikan bug)
  - `refactor:` (Refactoring tanpa mengubah fungsionalitas)
  - `test:` (Penambahan pengujian)
- Selalu branch dari `develop` (`feature/nama-fitur`).