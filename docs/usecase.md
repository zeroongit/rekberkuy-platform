# 📋 RekberKuy — Use Case Document

Dokumen ini menjabarkan seluruh skenario penggunaan (use case) platform RekberKuy dari sudut pandang setiap aktor yang terlibat.

---

## 👥 Aktor

| Aktor | Deskripsi |
|-------|-----------|
| **Pembeli** | Pengguna yang ingin membeli barang, memesan jasa, atau membeli tiket event |
| **Penjual** | Pengguna yang menjual barang fisik maupun digital |
| **Penyedia Jasa** | Pengguna yang menawarkan layanan profesional / freelance |
| **Event Organizer (EO)** | Pengguna yang membuat dan mengelola event |
| **Vendor** | Penyedia layanan pendukung event (katering, dekorasi, fotografer, dll) |
| **Admin** | Tim internal RekberKuy yang mengelola platform & mediasi sengketa |
| **Sistem** | Proses otomatis yang dijalankan platform (timer, notifikasi, blockchain logger) |

---

## 🔐 UC-01: Autentikasi & Manajemen Akun

### UC-01.1 — Daftar Akun Baru
**Aktor:** Pembeli / Penjual / EO / Vendor
**Prekondisi:** Pengguna belum memiliki akun

**Alur Normal:**
1. Pengguna membuka halaman register
2. Pengguna mengisi nama, email, nomor HP, dan password
3. Sistem mengirim email verifikasi
4. Pengguna mengklik link verifikasi
5. Akun aktif dan pengguna diarahkan ke dashboard

**Alur Alternatif:**
- Email sudah terdaftar → sistem tampilkan pesan error & tawarkan login
- Link verifikasi expired → pengguna bisa minta kirim ulang

---

### UC-01.2 — Login
**Aktor:** Semua pengguna terdaftar

**Alur Normal:**
1. Pengguna masukkan email & password
2. Sistem validasi kredensial via Supabase Auth
3. Sistem buat session token (JWT)
4. Pengguna diarahkan ke dashboard

**Alur Alternatif:**
- Salah password 3x → akun dikunci sementara 15 menit
- Lupa password → alur reset password via email

---

### UC-01.3 — Verifikasi Identitas (KYC)
**Aktor:** Penjual, EO, Vendor
**Prekondisi:** Akun sudah aktif, belum terverifikasi

**Alur Normal:**
1. Pengguna masuk ke menu KYC
2. Pengguna upload foto KTP/SIM (depan & belakang)
3. Pengguna upload foto selfie dengan KTP
4. Pengguna isi data diri sesuai KTP
5. Sistem mengirim dokumen ke Admin untuk review
6. Admin memverifikasi dalam 1×24 jam
7. Pengguna menerima notifikasi: verifikasi berhasil
8. Badge "Terverifikasi" muncul di profil pengguna

**Alur Alternatif:**
- Dokumen tidak jelas / tidak sesuai → Admin tolak & minta upload ulang
- Data tidak cocok dengan KTP → Admin tolak & pengguna bisa ajukan banding

---

## 💳 UC-02: Wallet & Pembayaran

### UC-02.1 — Top-Up Saldo Wallet
**Aktor:** Semua pengguna terdaftar

**Alur Normal:**
1. Pengguna masuk ke menu Wallet
2. Pengguna pilih nominal top-up
3. Pengguna pilih metode pembayaran (transfer bank / e-wallet)
4. Sistem buat tagihan via payment gateway (Midtrans)
5. Pengguna selesaikan pembayaran
6. Payment gateway kirim callback ke sistem
7. Sistem update saldo wallet pengguna
8. Pengguna terima notifikasi saldo berhasil ditambahkan

---

### UC-02.2 — Tarik Saldo (Withdraw)
**Aktor:** Penjual, Penyedia Jasa, EO, Vendor

**Alur Normal:**
1. Pengguna masuk ke menu Wallet → Tarik Dana.
2. Pengguna masukkan nominal & pilih rekening tujuan.
3. Sistem memvalidasi kelayakan saldo: **(Nominal Penarikan + Batas Flat Fee Rp 7.500) <= Total Saldo Aktif**.
4. Sistem memotong saldo wallet dan mencatat data mutasi ke tabel `wallet_ledger_logs` dengan tipe `KREDIT`.
5. Sistem memproses transfer ke rekening pengguna (H+1 kerja via Midtrans/Disbursement API).
6. Pengguna terima notifikasi transfer berhasil.

---

## 🛍️ UC-03: Transaksi Barang

### UC-03.1 — Buat Transaksi Baru (Inisiasi oleh Pembeli)
**Aktor:** Pembeli
**Prekondisi:** Pembeli memiliki saldo wallet yang cukup

**Alur Normal:**
1. Pembeli pilih "Buat Transaksi Barang"
2. Pembeli isi detail transaksi:
   - Nama barang & deskripsi
   - Harga barang
   - Data penjual (username atau email)
   - Metode pengiriman
3. Sistem hitung total (harga + ongkir + fee platform)
4. Pembeli konfirmasi & saldo terpotong → masuk ke escrow
5. Penjual menerima notifikasi transaksi baru
6. Sistem log transaksi ke blockchain Avalanche (audit)

**Alur Alternatif:**
- Saldo tidak cukup → sistem arahkan ke halaman top-up
- Data penjual tidak ditemukan → sistem tampilkan error

---

### UC-03.2 — Konfirmasi & Pengiriman oleh Penjual
**Aktor:** Penjual
**Prekondisi:** Transaksi sudah dibuat, dana sudah di escrow

**Alur Normal:**
1. Penjual terima notifikasi transaksi masuk
2. Penjual konfirmasi kesanggupan mengirim barang
3. Penjual input nomor resi pengiriman
4. Sistem simpan resi & update status → "Dalam Pengiriman"
5. Pembeli terima notifikasi barang dikirim

---

### UC-03.3 — Konfirmasi Penerimaan oleh Pembeli
**Aktor:** Pembeli
**Prekondisi:** Penjual sudah input nomor resi

**Alur Normal:**
1. Pembeli menerima barang
2. Pembeli buka aplikasi → konfirmasi "Barang Diterima"
3. Sistem release dana escrow ke wallet penjual
4. Penjual terima notifikasi dana masuk
5. Sistem update status transaksi → "Selesai"
6. Sistem update audit log on-chain → status final

**Alur Alternatif (Timeout):**
- Pembeli tidak konfirmasi dalam 3 hari setelah estimasi tiba → sistem otomatis release dana ke penjual

---

### UC-03.4 — Buka Sengketa (Dispute) & Rating Rendah
**Aktor:** Pembeli, Sistem (AI Multimodal Engine), Admin
**Prekondisi:** Transaksi masih berjalan, atau pembeli ingin memberikan rating < 3 bintang.

**Alur Normal:**
1. Pembeli pilih "Laporkan Masalah / Beri Rating Rendah" di halaman transaksi.
2. Sistem mendeteksi rating < 3 / klaim cacat, lalu memunculkan hard-constraint: **Wajib Unggah Video Unboxing**.
3. Pembeli mengunggah bukti video unboxing tanpa terputus (*uncut*).
4. Sistem memicu *AI Claim Verifier* (Gemini Multimodal API) untuk membedah integritas video, cek manipulasi frame, dan OCR nomor resi.
5. Jika AI menyatakan `VALID_CLAIM` dengan confidence score > 0.8:
   - Sistem melakukan freeze dana di escrow.
   - Notifikasi dikirim ke penjual dan admin masuk sebagai mediator.
6. Admin review kesimpulan AI dan memutuskan resolusi akhir.
7. Sistem eksekusi keputusan & update audit log on-chain Avalanche.

**Alur Alternatif:**
- Pembeli tidak mengunggah video / video terdeteksi hasil editan (cuts) -> Sistem otomatis **menolak sengketa/ulasan** sebelum masuk database. Status ulasan ditandai sebagai `FILTERED`.

---

## 🛠️ UC-04: Transaksi Jasa

### UC-04.1 — Buat Transaksi Jasa
**Aktor:** Pembeli (Klien)
**Prekondisi:** Klien & Penyedia Jasa sudah sepakat di luar platform

**Alur Normal:**
1. Klien pilih "Buat Transaksi Jasa"
2. Klien isi detail:
   - Nama & deskripsi pekerjaan
   - Data penyedia jasa
   - Total nilai kontrak
   - Daftar milestone (nama, nominal, deadline)
3. Penyedia jasa konfirmasi setuju dengan detail
4. Klien deposit total dana ke escrow
5. Transaksi aktif, milestone pertama dimulai

---

### UC-04.2 — Penyelesaian Milestone
**Aktor:** Penyedia Jasa, Klien
**Prekondisi:** Transaksi jasa aktif

**Alur Normal:**
1. Penyedia jasa tandai milestone sebagai selesai
2. Penyedia jasa upload hasil kerja / bukti penyelesaian
3. Klien menerima notifikasi untuk review
4. Klien review hasil kerja
5. Klien konfirmasi milestone selesai
6. Sistem release dana milestone ke wallet penyedia jasa
7. Lanjut ke milestone berikutnya

**Alur Alternatif:**
- Klien minta revisi → penyedia jasa upload ulang
- Klien tidak respons dalam 3 hari → dana milestone otomatis release

---

## 🎪 UC-05: Transaksi Event

### UC-05.1 — Buat Event (oleh EO)
**Aktor:** Event Organizer (EO)
**Prekondisi:** EO sudah terverifikasi (KYC)

**Alur Normal:**
1. EO pilih "Buat Event Baru"
2. EO isi informasi event:
   - Nama, deskripsi, tanggal & lokasi
   - Kategori event (konser, seminar, pesta, dll)
   - Kapasitas & harga tiket
3. EO kelola kebutuhan vendor:
   - Pilih vendor dari marketplace RekberKuy, ATAU
   - Daftarkan vendor eksternal (nama, kontak, nilai kontrak)
4. EO submit event untuk review Admin (opsional)
5. Event dipublikasikan / siap menerima peserta

---

### UC-05.2 — Pembelian Tiket (oleh Peserta)
**Aktor:** Pembeli (Peserta Event)
**Prekondisi:** Event sudah dipublikasikan, tiket tersedia

**Alur Normal:**
1. Peserta browse atau cari event
2. Peserta pilih jumlah & kategori tiket
3. Sistem hitung total + fee platform
4. Peserta bayar via wallet / payment gateway
5. Sistem generate tiket digital (QR Code)
6. Peserta terima tiket via email & in-app
7. Dana tiket masuk ke escrow event

---

### UC-05.3 — Pembayaran Vendor oleh EO
**Aktor:** EO, Vendor
**Prekondisi:** Vendor sudah dikontrak untuk event

**Alur Normal:**
1. EO buat kontrak vendor dari halaman event
2. EO tentukan milestone pembayaran vendor
3. EO deposit dana vendor ke escrow
4. Vendor selesaikan pekerjaan sesuai milestone
5. EO konfirmasi pekerjaan vendor selesai
6. Sistem release dana ke wallet vendor
7. Sistem log pembayaran vendor ke blockchain

---

### UC-05.4 — Penyelesaian Event & Release Dana Tiket
**Aktor:** EO, Sistem

**Alur Normal:**
1. EO tandai event sebagai "Selesai".
2. Sistem mengumpulkan seluruh data pengeluaran dari tabel `event_vendor_allocations` dan total pendapatan tiket.
3. Sistem mengeksekusi usecase `finance_calculator.go` untuk mendapatkan output struktur data `EventAuditResult`.
4. Sistem mendistribusikan dana secara otomatis:
   - Platform Fee ditarik ke dompet pendapatan sistem.
   - Alokasi dana bersih sub-vendor dikirim ke dompet masing-masing vendor mitra.
   - Bonus surplus bersih (jika ada) disuntikkan ke wallet EO bersama Management Fee mereka.
5. Sistem memperbarui status transaksi menjadi `RELEASED` dan mencatat transaksi akhir tersebut ke blockchain Avalanche.

**Alur Alternatif (Event Dibatalkan):**
1. EO batalkan event
2. Sistem otomatis refund semua uang ke wallet peserta
3. Kontrak vendor diputuskan — dana dikembalikan ke EO

---

## 🏪 UC-06: Vendor Marketplace

### UC-06.1 — Pendaftaran Vendor
**Aktor:** Vendor
**Prekondisi:** Vendor punya akun RekberKuy

**Alur Normal:**
1. Vendor masuk ke "Daftar sebagai Vendor"
2. Vendor isi profil bisnis:
   - Nama usaha & deskripsi layanan
   - Kategori (katering, dekorasi, foto/video, sound system, dll)
   - Area layanan (kota/provinsi)
   - Portofolio (foto/dokumen)
   - Harga mulai dari
3. Vendor upload dokumen legalitas (SIUP, NPWP, dll)
4. Admin verifikasi dalam 2×24 jam
5. Vendor tampil di marketplace dengan badge "Verified Vendor"

---

### UC-06.2 — EO Memilih Vendor dari Marketplace
**Aktor:** EO
**Prekondisi:** Event sudah dibuat

**Alur Normal:**
1. EO buka tab "Cari Vendor" di halaman event
2. EO filter berdasarkan kategori, kota, & anggaran
3. EO lihat profil & portofolio vendor
4. EO hubungi vendor via platform (chat)
5. EO & Vendor sepakat soal detail layanan & harga
6. EO buat kontrak vendor → dana masuk escrow
7. Vendor terima notifikasi kontrak baru

---

## ⚙️ UC-07: Admin

### UC-07.1 — Moderasi & Manajemen Platform
**Aktor:** Admin

- Review & approve/reject KYC pengguna
- Review & approve/reject pendaftaran vendor
- Mediasi sengketa transaksi
- Pantau semua transaksi aktif di dashboard
- Suspend / ban akun yang melanggar aturan
- Kelola fee & konfigurasi platform

---

## 🔗 Diagram Alur Transaksi Barang

```
Pembeli                  Sistem                   Penjual
   │                        │                        │
   ├── Buat Transaksi ──────►│                        │
   │                        ├── Potong Saldo ─────── │
   │                        ├── Dana ke Escrow ────── │
   │                        ├── Notif Penjual ───────►│
   │                        │                        │
   │                        │◄── Konfirmasi Kirim ───┤
   │                        ├── Update Status ─────── │
   │◄── Notif Dikirim ──────┤                        │
   │                        │                        │
   ├── Konfirmasi Terima ───►│                        │
   │                        ├── Release Escrow ─────► │
   │                        ├── Log Blockchain ─────── │
   │                        ├── Notif Penjual ───────►│
   │◄── Notif Selesai ──────┤                        │
```

---

## 🔗 Diagram Alur Transaksi Event

```
EO                       Sistem                   Peserta / Vendor
   │                        │                        │
   ├── Buat Event ─────────►│                        │
   ├── Setup Vendor ────────►│                        │
   │                        ├── Publikasi Event ─────►│ (Peserta)
   │                        │◄── Beli Tiket ─────────┤
   │                        ├── Dana Tiket Escrow ─── │
   │                        │                        │
   ├── Kontrak Vendor ──────►│                        │
   │                        ├── Dana Vendor Escrow───►│ (Vendor)
   │                        │◄── Kerja Selesai ───────┤
   ├── Konfirmasi Vendor ───►│                        │
   │                        ├── Release ke Vendor ───►│
   │                        │                        │
   ├── Event Selesai ───────►│                        │
   │                        ├── Fee Platform ─────── │
   │◄── Release Tiket ──────┤                        │
   │                        ├── Log Blockchain ─────── │
```