# ⚙️ RekberKuy - Core Service (Backend)

Modul ini adalah backend utama untuk platform **RekberKuy**, dibangun menggunakan **Go 1.25** dengan menerapkan pola **Clean Architecture** murni. Backend ini mengelola seluruh logika bisnis, transaksi escrow (Barang, Jasa, Event), manajemen pengguna (CRM Loyalty), hingga fungsi *relayer* untuk pencatatan *gasless transaction* ke blockchain.

## 🏗️ Arsitektur (Clean Architecture)
Sistem ini sangat ketat dalam memisahkan *concern* melalui lapisan hierarki berikut:

1. **`delivery/http/`**: Handler REST API (Gin/Echo/Chi). Bertugas memproses request HTTP, validasi payload awal, dan mereturn response JSON.
2. **`usecase/`**: Berisi murni logika bisnis (contoh: alur escrow, `finance_calculator`).
3. **`repository/`**: Layer akses ke basis data (PostgreSQL via Supabase & Redis).
4. **`domain/`**: Definisi *struct* (Entity) dan *interface* (Contract) yang independen.

> **Aturan Emas:** *Layer* luar hanya boleh memanggil *layer* yang lebih dalam. `domain` tidak boleh mengimpor paket dari *layer* manapun.

## 💸 Konteks Transaksi
- **Murni IDR**: Backend ini mengelola keuangan dalam bentuk mata uang Rupiah. Tidak ada penyimpanan alamat kripto milik pengguna.
- **Gasless Relayer**: Backend bertugas menembak *smart contract* ke Avalanche secara otomatis di latar belakang (*background job*). Pengguna akhir tidak dibebani *gas fee*.
- **ACID Compliance**: Semua mutasi dompet/keuangan wajib menggunakan *database transaction* dan *pessimistic locking* (`SELECT ... FOR UPDATE`).

## 🚀 Cara Menjalankan Lokal

1. **Persiapan Lingkungan**: Pastikan `.env` sudah dikonfigurasi berdasarkan `.env.example`.
2. **Install Dependensi**:
   ```bash
   go mod tidy
   ```
3. **Jalankan Server**:
   ```bash
   go run ./cmd/server/main.go
   ```
4. **Build Binary**:
   ```bash
   go build -o bin/server ./cmd/server/main.go
   ```

## 🧪 Pengujian
Setiap penambahan *usecase* baru wajib menyertakan unit test.
```bash
# Menjalankan semua test
go test ./...
```