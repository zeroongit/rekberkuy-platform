# 🤖 Claude AI - Core Service (Backend)

Panduan lokal ini memberikan konteks khusus untuk pengembangan backend **RekberKuy**. Pastikan selalu mematuhi aturan arsitektur ini sebelum menulis kode.

## 🏗️ Clean Architecture Murni
Arsitektur sangat ketat, layer tidak boleh dilompati atau mengimpor layer luar secara terbalik:
- **`delivery/http/`**: Handler HTTP murni. Hanya untuk routing, parsing request/response. Dilarang menaruh logika bisnis. (Boleh import `usecase/` & `domain/`)
- **`usecase/`**: Inti logika bisnis. Menghubungkan repository dan mengeksekusi kalkulasi (seperti `finance_calculator`). (Boleh import `domain/`)
- **`repository/`**: Layer akses data (PostgreSQL via Supabase & Redis). Mengimplementasikan interface dari domain. (Boleh import `domain/`)
- **`domain/`**: Entity utama & Interface (contract). **DILARANG MUTLAK** mengimport layer lain.

## 💰 Konteks Bisnis & Gasless Transaction
- Sistem **murni IDR (Rupiah)**. Dilarang meminta atau menyimpan crypto address pengguna.
- Pencatatan ke blockchain memanfaatkan metode **Gasless Transaction**. Backend bertindak sebagai *Relayer* yang mengeksekusi smart contract secara otomatis di latar belakang.

## 🔐 Aturan Mutasi & Transaksi Database
- Semua operasi yang menyangkut saldo (`wallet_repository.go` dsb) **WAJIB** dibungkus dalam Database Transaction (`BeginTx`).
- Gunakan proteksi *Anti Race-Condition* dengan `SELECT ... FOR UPDATE` (Pessimistic Locking) sebelum melakukan mutasi saldo untuk menghindari *double spending*.

## 📐 Konvensi Kode (Golang)
- **Nama File**: `snake_case.go` (contoh: `transaction_usecase.go`).
- **Penamaan Struct/Func**: `PascalCase` untuk yang diekspor, `camelCase` untuk private.
- **Context**: Parameter pertama dari fungsi I/O (usecase, repository) wajib `ctx context.Context`.
- **Error Handling**: Selalu kembalikan `error`, dilarang keras menggunakan `panic()`.

## 🛠️ Command Penting
```bash
# Install dependencies
go mod tidy

# Jalankan server
go run ./cmd/server/main.go

# Jalankan test
go test ./...
```