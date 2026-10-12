# Business Concept & Escrow Monetization Model — RekberKuy

## 1. Konsep Bisnis Inti
RekberKuy adalah platform rekening bersama (escrow) digital yang dirancang untuk menjembatani trust gap dalam transaksi peer-to-peer dan B2B di Indonesia. Berbeda dengan escrow konvensional manual (yang mengandalkan transfer ke rekening pribadi admin grup media sosial atau forum yang rentan penipuan), RekberKuy menyediakan ekosistem terstruktur dengan sistem penahanan dana otomatis (escrow), verifikasi identitas (KYC), deteksi risiko berbasis AI, mediasi sengketa transparan, dan rekam jejak audit on-chain di blockchain Avalanche (gasless).

Tiga pilar vertikal utama platform:
1. **Goods (Barang):** Jual beli barang fisik/digital (elektronik, hobi, fashion, dll.) dengan proteksi resi dan auto-release 3 hari.
2. **Services (Jasa):** Layanan profesional & freelancer dengan sistem pembayaran berbasis milestone bertahap.
3. **Events (Event & Vendor):** Pengadaan event, penjualan tiket ber-QR code, serta marketplace vendor (katering, dekorasi, EO) dengan manajemen keuangan terpusat.

---

## 2. Model Monetisasi & Revenue Streams

RekberKuy mengandalkan model monetisasi berkelanjutan yang transparan dan adil bagi seluruh pihak:

### A. Komisi Platform dari Transaksi (Take Rate)
- **Transaksi Barang:** Fee platform sebesar **1.5%** dari nilai transaksi (dibebankan kepada pembeli/penjual sesuai skema tier).
- **Transaksi Jasa:** Fee platform sebesar **2.0%** dari nilai kontrak (dipotong saat release akhir).
- **Tiket Event:** Fee platform sebesar **2.0%** dari total penjualan tiket.
- **Kontrak Vendor Event:** Fee platform sebesar **1.5%** dari nilai kontrak vendor.

### B. Buyer Protection Fee
- Biaya proteksi tambahan berdasarkan tier penjual (Newbies: Rp 2.500 via RekberPay / Rp 5.000 via payment gateway; Silver: 4%; Gold: 8%).

### C. Biaya Tarik Saldo (Withdrawal Fee)
- Penarikan saldo RekberPay di bawah Rp 1.000.000 dikenakan biaya flat Rp 2.500 — Rp 7.500 untuk menutupi biaya interbank clearing (T+1). Penarikan di atas Rp 1.000.000 gratis.

### D. CRM Loyalty Tiering Benefits
- Pengguna dengan volume transaksi tinggi masuk ke tier Silver, Gold, atau Platinum, mendapatkan potongan fee (diskon take rate hingga 0.5%) dan prioritas mediasi sengketa.

---

## 3. Aliran Dana & Keamanan Keuangan (Escrow Logic)
- **Segregated Pool:** Dana pembeli yang masuk melalui Midtrans disimpan dalam rekening penampungan terpusat (escrow pool) dan dicatat secara akuntansi ganda (double-entry ledger) di tabel `rekberpay_transactions`.
- **Zero Crypto Exposure:** Pengguna tidak pernah memegang crypto atau membayar gas fee. Seluruh pencatatan blockchain dilakukan oleh backend relayer secara gasless.
- **ACID Transactions:** Setiap mutasi dompet dan escrow dijamin konsistensinya menggunakan database transaction dengan `SELECT FOR UPDATE` untuk mencegah race condition atau double spending.
