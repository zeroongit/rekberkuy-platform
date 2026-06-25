# 🧠 Gemini Code Assist - Core Service (Backend)

Dokumentasi lokal ini dirancang untuk memberikan konteks kepada AI assistant mengenai arsitektur, konvensi, dan aturan main di backend **RekberKuy**.

## 🏗️ Arsitektur: Clean Architecture
Backend ini menggunakan Golang 1.25 dengan pola **Clean Architecture** murni. Layer berurutan dan memiliki hierarki ketergantungan (Dependency Rule) yang ketat:

1. **`delivery/http/`**: Handler HTTP. Hanya bertugas parsing request (JSON/Query) dan membungkus response. Tidak boleh ada logika bisnis di sini. (Boleh import `usecase/` dan `domain/`)
2. **`usecase/`**: Inti logika bisnis. Menghubungkan berbagai repository, mengeksekusi kalkulasi (seperti `finance_calculator`), dan mengatur alur transaksi. Tidak boleh import package HTTP. (Boleh import `domain/`)
3. **`repository/`**: Akses database ke PostgreSQL (via Supabase) dan Redis. Mengimplementasikan interface dari domain. (Boleh import `domain/`)
4. **`domain/`**: Entity utama (structs) dan interface (contracts). Bersifat independen, **tidak boleh import layer lain**. Sistem *murni IDR*. Integrasi *audit log* harus menggunakan metode *Gasless Transaction* (backend berperan sebagai relayer).

**PENTING:** Jangan pernah menembus layer (contoh: `delivery` langsung memanggil `repository`, atau `domain` mengimport `usecase`).

## 🔐 Aturan Transaksi Database (ACID)
- Semua mutasi yang melibatkan uang (seperti di `wallet_repository.go`) **WAJIB** menggunakan database transaction (`BeginTx`).
- Gunakan level isolasi `sql.LevelSerializable` jika memungkinkan.
- Selalu gunakan `SELECT ... FOR UPDATE` saat membaca saldo sebelum diubah untuk menghindari kondisi *Race Condition* atau *Double Spending*.

## 📂 Domain Bisnis Utama
- **User**: Entity pengguna & CRM Loyalty Tiering (target GMV/streak).
- **Transaction**: Transaksi escrow utama (Barang, Jasa, Event) & state status.
- **Finance**: Penampung hasil kalkulasi audit, fee platform, bonus, dan auto-refund.
- **Vendor**: Profil vendor mitra & skema alokasi anggaran event secara dinamis.
- **Wallet**: Sistem dompet (RekberPay), mutasi top-up, withdraw, lock funds, dan log transaksi logistik.

## 🛠️ Perintah Berguna
```bash
# Pindah ke direktori
cd apps/core-service

# Install / Rapihkan dependencies
go mod tidy

# Menjalankan server lokal (pastikan .env sudah diatur)
go run ./cmd/server/main.go

# Menjalankan unit tests
go test ./...

# Build binary
go build -o bin/server ./cmd/server/main.go
```

## 📐 Konvensi Kode
- **Penamaan File**: `snake_case` (contoh: `transaction_usecase.go`).
- **Penamaan Fungsi/Struct**: `PascalCase` untuk yang di-export, `camelCase` untuk yang private.
- **Error Handling**: Selalu kembalikan `error`, jangan gunakan `panic()` (kecuali di inisialisasi awal di `main.go`).
- **Context**: Gunakan `context.Context` sebagai argumen pertama pada fungsi layer `usecase` dan `repository`.