# Functional Requirements: Autentikasi & RBAC (FR-AUTH)

## 1. Overview
Modul autentikasi mengatur pendaftaran pengguna, login, manajemen sesi berbasis JWT, pemulihan kata sandi, serta kontrol akses berbasis peran (RBAC) yang memisahkan secara tegas antara **Akun Personal (Konsumen Ritel)** dan **Akun Komersial (Pelaku Bisnis/Merchant)**.

## 2. Matriks Pembatasan Hak Akses (Personal vs Komersial)

| Fitur / Aksi | Akun Personal (`USER`) | Akun Komersial (`SELLER`, `VERIFIED_MERCHANT`, `VERIFIED_VENDOR`, `EVENT_ORGANIZER`, `SERVICE_PROVIDER`) |
|---|---|---|
| **Mendaftar & Login** | Diizinkan | Diizinkan (memerlukan verifikasi KYC/Dokumen Legalitas) |
| **Membeli Barang / Jasa** | Diizinkan | Diizinkan (hanya untuk keperluan operasional bisnis internal) |
| **Membeli Tiket Event** | Diizinkan (sebagai penonton/peserta reguler) | **Dilarang Keras (403 Forbidden)** — entitas komersial tidak diizinkan membeli tiket event |
| **Berjualan Produk / Jasa** | Tidak diizinkan | Diizinkan sesuai domain bisnis masing-masing |
| **Membuat Event / Vendor** | Tidak diizinkan | Diizinkan (khusus Event Organizer / Vendor) |

## 3. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-AUTH-01** | Pengguna dapat mendaftar dengan email, nama, nomor HP, dan password. | M | Implemented (`user_usecase.go`) |
| **FR-AUTH-02** | Sistem mengirim email verifikasi setelah pendaftaran berhasil. | M | Implemented (SMTP Adapter) |
| **FR-AUTH-03** | Pengguna dapat login dengan email dan password (hashing bcrypt cost factor 12). | M | Implemented (`auth_usecase.go`) |
| **FR-AUTH-04** | Sistem menerbitkan JWT dengan expiry 24 jam disimpan sebagai HttpOnly, Secure, SameSite=Strict cookie. | M | Implemented (`auth_middleware.go`) |
| **FR-AUTH-05** | Pengguna dapat melakukan logout — menghapus cookie dan sesi aktif. | M | Implemented |
| **FR-AUTH-06** | Pengguna dapat melakukan reset password melalui tautan verifikasi email. | M | Implemented |
| **FR-AUTH-07** | Sistem mengunci akun secara otomatis selama 15 menit setelah 5 kali percobaan login gagal berturut-turut. | M | Implemented |
| **FR-AUTH-08** | Semua endpoint protected memerlukan token JWT yang valid dan pengecekan role yang sesuai. | M | Implemented |
| **FR-AUTH-08B** | Endpoint pembelian tiket event wajib memvalidasi dan menolak akses jika request berasal dari akun komersial (`EVENT_ORGANIZER`, `VENDOR`, `SELLER`, `SERVICE_PROVIDER`) menggunakan middleware `RequirePersonalAccount()`. | M | Implemented (`auth_middleware.go`) |
| **FR-AUTH-09** | Pembaruan token otomatis (refresh token mechanism) jika token mendekati masa kadaluarsa. | S | Planned / Partial |
| **FR-AUTH-10** | Opsi login alternatif menggunakan Google OAuth. | C | Planned (Phase 3) |
