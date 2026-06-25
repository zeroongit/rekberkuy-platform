# 🗺️ RekberKuy — Product Roadmap

Dokumen ini menggambarkan rencana pengembangan platform RekberKuy secara bertahap. Setiap fase memiliki tujuan yang jelas dan deliverable yang terukur.

---

## 🎯 Visi Produk

Menjadi platform rekening bersama (escrow) **paling dipercaya di Indonesia** yang melayani transaksi barang, jasa, dan event dalam satu ekosistem yang transparan, aman, dan efisien.

---

## 📊 Status Saat Ini

```
Phase 0 ████████░░░░░░░░░░░░  40% — Foundation & Setup
Phase 1 ░░░░░░░░░░░░░░░░░░░░   0% — Belum dimulai
Phase 2 ░░░░░░░░░░░░░░░░░░░░   0% — Belum dimulai
Phase 3 ░░░░░░░░░░░░░░░░░░░░   0% — Belum dimulai
Phase 4 ░░░░░░░░░░░░░░░░░░░░   0% — Belum dimulai
```

---

## 🏗️ Phase 0 — Foundation (Sekarang)
**Target:** Setup infrastruktur & arsitektur solid sebelum membangun fitur

### ✅ Selesai
- [x] Inisialisasi monorepo (`apps/`, `blockchain/`, `e2e-qa/`, `docs/`)
- [x] Setup Next.js v16 + Shadcn/UI + Tailwind CSS v4
- [x] Setup Go backend dengan clean architecture
- [x] Domain model: `user`, `transaction`, `finance`, `vendor`, `wallet`
- [x] Usecase: `finance_calculator`, `transaction_usecase`
- [x] Setup Hardhat v3 + smart contract dasar
- [x] Struktur `e2e-qa/` untuk QA Automation
- [x] GitLab CI/CD pipeline
- [x] Dokumentasi: README, CLAUDE.md, usecase, concept, roadmap

### 🔄 In Progress
- [ ] HTTP handler delivery layer (routing semua domain)
- [ ] Repository lengkap untuk semua domain
- [ ] Konfigurasi Supabase + migrasi database awal
- [ ] Environment setup untuk staging & production

---

## 🚀 Phase 1 — MVP: Transaksi Barang
**Target Durasi:** 6–8 minggu
**Tujuan:** Menghasilkan produk yang bisa dipakai oleh pengguna nyata untuk transaksi barang

### Backend
- [ ] Auth API — register, login, logout, refresh token (via Supabase)
- [ ] User API — profil, verifikasi email
- [ ] KYC API — upload dokumen, verifikasi identitas penjual & EO
- [ ] Wallet API — saldo, top-up, riwayat
- [ ] Transaction API (Barang) — buat, konfirmasi, batalkan, dispute
- [ ] Finance API — kalkulasi fee platform & pajak
- [ ] Middleware — auth JWT, rate limiting, logging

### Frontend
- [ ] Halaman landing page (marketing)
- [ ] Halaman auth — login & register
- [ ] Dashboard utama — ringkasan transaksi & saldo wallet
- [ ] Alur transaksi barang end-to-end:
  - Form buat transaksi baru
  - Halaman detail transaksi
  - Konfirmasi pembayaran
  - Konfirmasi penerimaan barang
  - Pelepasan dana ke penjual
- [ ] Halaman riwayat transaksi

### Infrastruktur
- [ ] Integrasi payment gateway (Midtrans) untuk top-up wallet
- [ ] Setup Redis untuk caching sesi & data sering diakses
- [ ] Email notifikasi (SMTP) untuk event transaksi penting
- [ ] Deployment staging environment

### Testing
- [ ] Unit test semua usecase backend
- [ ] Skenario E2E alur transaksi barang di `e2e-qa/`
- [ ] Postman collection untuk semua endpoint Phase 1

### 🎯 Definition of Done Phase 1
> Pengguna bisa mendaftar, top-up wallet, membuat transaksi barang dengan penjual, mengkonfirmasi penerimaan, dan dana berhasil dikirim ke penjual.

---

## ⚡ Phase 2 — Modul Jasa & Event Dasar
**Target Durasi:** 6–8 minggu
**Tujuan:** Ekspansi ke domain jasa dan event, lengkap dengan fitur KYC

### Backend
- [ ] Transaction API (Jasa) — milestone-based payment
- [ ] Event API — buat event, manajemen tiket, status event
- [ ] Dispute & Mediasi API — buka sengketa, upload bukti, resolusi admin
- [ ] Notifikasi real-time via Supabase Realtime
- [ ] AI Multimodal Review Guard: Integrasi Gemini API untuk membedah video unboxing (uncut verification) saat pembeli memberikan rating < 3 bintang.

### Frontend
- [ ] Alur transaksi jasa end-to-end:
  - Posting kebutuhan jasa
  - Penawaran dari penyedia jasa
  - Persetujuan milestone
  - Konfirmasi penyelesaian per milestone
- [ ] Alur pembuatan event (EO):
  - Form buat event
  - Manajemen detail event
  - Sistem tiket & kapasitas
- [ ] Halaman KYC — upload & status verifikasi
- [ ] Sistem dispute — form laporan, status, chat mediasi
- [ ] Notifikasi in-app (bell icon + daftar notifikasi)

### Infrastruktur
- [ ] Storage Supabase untuk upload dokumen KYC & bukti dispute
- [ ] Push notification (Web Push API)
- [ ] Audit log — setiap event transaksi dikirim ke smart contract Avalanche (Fuji Testnet)

### Testing
- [ ] Unit test modul jasa & event
- [ ] Skenario E2E alur jasa (milestone) & alur event
- [ ] Load test endpoint transaksi (Postman + k6)

### 🎯 Definition of Done Phase 2
> Penjual jasa bisa menerima pembayaran bertahap. EO bisa membuat event dan menjual tiket. Pengguna terverifikasi (KYC). Setiap transaksi tercatat di blockchain Avalanche testnet.

---

## 🤝 Phase 3 — Vendor Marketplace
**Target Durasi:** 4–6 minggu
**Tujuan:** Mengaktifkan ekosistem vendor untuk EO di dalam platform

### Konsep
EO yang mengadakan event membutuhkan vendor (katering, dekorasi, fotografer, dll). Platform menyediakan dua opsi:
1. **Vendor Rekanan RekberKuy** — vendor yang sudah terverifikasi di platform
2. **Vendor Eksternal** — vendor pilihan EO sendiri, tapi pembayarannya tetap via escrow RekberKuy

### Backend
- [ ] Vendor API — pendaftaran, profil, kategori, portofolio
- [ ] Vendor Onboarding — verifikasi vendor, persetujuan admin
- [ ] Vendor Catalog API — pencarian, filter kategori, rating
- [ ] Kontrak Vendor API — buat kontrak EO-vendor, milestone pembayaran

### Frontend
- [ ] Halaman marketplace vendor (publik) — browse & cari vendor
- [ ] Halaman profil vendor — portofolio, layanan, harga, rating
- [ ] Alur onboarding vendor baru — form pendaftaran & dokumen
- [ ] Integrasi di flow event — EO pilih vendor saat setup event
- [ ] Manajemen kontrak vendor — status, milestone, pembayaran

### Infrastruktur
- [ ] Search & filter vendor (PostgreSQL full-text search atau Elasticsearch)
- [ ] Rating & review system

### 🎯 Definition of Done Phase 3
> EO bisa browse vendor di platform, memilih vendor rekanan atau daftarkan vendor eksternal, membuat kontrak dengan escrow payment, dan vendor menerima pembayaran setelah event selesai.

---

## 🌟 Phase 4 — Blockchain Production & Skala
**Target Durasi:** 4–6 minggu
**Tujuan:** Go live di Avalanche Mainnet, optimasi performa, dan fitur lanjutan

### Blockchain
- [ ] Audit & security review smart contract
- [ ] Deploy `TransactionLogger` ke Avalanche C-Chain Mainnet
- [ ] Explorer halaman publik — siapapun bisa verifikasi transaksi via hash
- [ ] Integrasi penuh dari backend — setiap transaksi selesai → catat on-chain

### Platform
- [ ] Affiliate & referral program
- [ ] Integrasi e-wallet (GoPay, OVO, Dana) untuk top-up
- [ ] Laporan & analytics dashboard untuk admin
- [ ] Multi-bahasa (Bahasa Indonesia & English)
- [ ] Dark mode UI
- [ ] PWA (Progressive Web App)

### Keamanan & Compliance
- [ ] Penetration testing
- [ ] OWASP security audit
- [ ] Kepatuhan regulasi OJK (jika relevan)
- [ ] Enkripsi data sensitif at-rest

### 🎯 Definition of Done Phase 4
> Platform berjalan stabil di production. Setiap transaksi tercatat di Avalanche Mainnet. Pengguna bisa verifikasi independen transaksi mereka.

---

## 📱 Phase 5 — Mobile & Growth (Future)
**Target Durasi:** TBD
**Tujuan:** Ekspansi ke mobile dan fitur-fitur pertumbuhan

- [ ] Mobile app (React Native / Flutter)
- [ ] Fitur chat/negosiasi langsung antar pengguna
- [ ] AI-powered fraud detection
- [ ] Open API untuk integrasi pihak ketiga
- [ ] Program mitra bisnis (white-label escrow)
- [ ] Laporan pajak otomatis (e-Faktur)

---


---

## 📌 Prinsip Pengembangan

1. **Ship fast, iterate** — rilis fitur kecil dan sering, dapatkan feedback nyata
2. **Test before merge** — tidak ada kode masuk ke `develop` tanpa test
3. **Security first** — setiap endpoint butuh auth, setiap input butuh validasi
4. **Domain-driven** — kode diorganisir berdasarkan domain bisnis, bukan layer teknis
5. **Blockchain is audit only** — jangan tambahkan logika bisnis ke smart contract

---

*Dokumen ini bersifat living document — update setiap ada perubahan prioritas atau selesainya sebuah fase.*