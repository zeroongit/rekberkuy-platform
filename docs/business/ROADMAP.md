# Product Roadmap — RekberKuy Platform

## Overview
Roadmap pengembangan RekberKuy terbagi menjadi 4 fase utama, dimulai dari Minimum Viable Product (MVP) untuk transaksi barang peer-to-peer, ekspansi ke layanan dan event procurement, peluncuran vendor marketplace, hingga integrasi penuh on-chain audit log di Avalanche Mainnet.

---

## Phase 1 — MVP Barang (Q4 2026)
*Fokus Utama: Membangun fondasi escrow aman untuk transaksi barang fisik & digital.*

- **Autonomus Auth & Profile:** Registrasi lokal email/password, JWT HttpOnly cookie, role-based access.
- **Wallet & Payment:** Integrasi Midtrans Snap (VA, QRIS, Transfer Bank), dompet RekberPay, ledger mutasi saldo (ACID + FOR UPDATE).
- **Escrow Barang:** Pembuatan transaksi barang, fund locking, konfirmasi pengiriman resi oleh penjual, konfirmasi penerimaan oleh pembeli, auto-release 3 hari.
- **KYC Pre-screening:** Upload KTP & selfie, integrasi AI scoring (Groq) untuk pre-screening, review admin manual.
- **Dispute & Mediasi:** Mekanisme buka sengketa, freeze dana escrow, admin mediation panel (`REFUND_BUYER` / `RELEASE_TO_SELLER`).
- **Audit Log Relayer:** Integrasi relayer Go ke Avalanche Fuji Testnet (`TransactionLogger.sol`).
- **CRM Loyalty Tiering:** Evaluasi tier bulanan (Bronze, Silver, Gold, Platinum).

---

## Phase 2 — Jasa & Event Dasar (Q1 2027)
*Fokus Utama: Ekspansi ke layanan profesional (freelance) dan manajemen event.*

- **Transaksi Jasa (Milestone-based):** Kontrak jasa dengan tahapan milestone, lock total dana di muka, release bertahap per milestone, revisi deliverable.
- **Event Management & Ticketing:** Pembuatan event, manajemen tiket digital dengan QR code unik, escrow dana tiket, refund otomatis jika event batal.
- **AI Fraud & Risk Scoring:** Penyempurnaan fraud detection engine di `backend-ai` untuk mendeteksi anomali transaksi jasa dan event.
- **Notification Engine:** Email & in-app real-time notifications via Supabase Realtime / SMTP.

---

## Phase 3 — Vendor Marketplace (Q1 2027)
*Fokus Utama: Membangun ekosistem B2B antara Event Organizer dan Vendor Pendukung.*

- **Vendor Onboarding & Verification:** Pendaftaran vendor (katering, dekorasi, fotografer, sound system), upload legalitas (SIUP/NPWP), verifikasi admin.
- **Vendor Marketplace Directory:** Halaman direktori vendor dengan filter kategori, kota, dan rating.
- **EO-Vendor Contracts:** Integrasi kontrak vendor dari dalam halaman event, escrow pembayaran vendor, payout internal (wallet) dan eksternal (`PENDING_DISBURSEMENT` / Admin-triggered disbursement).
- **Event Finance Audit:** Kalkulator keuangan event otomatis (fee platform 5%, surplus distribution, bonus EO).

---

## Phase 4 — Avalanche Mainnet & Scale (Q2 2027)
*Fokus Utama: Migrasi ke production network dan scale-up operasional.*

- **Avalanche C-Chain Mainnet:** Deploy smart contract `TransactionLogger.sol` ke mainnet dengan cold wallet (owner) dan hot wallet (relayer).
- **Audit Reconciler Automation:** Peningkatan background worker untuk audit log reconciliation (reconciler 6h cycle).
- **Performance Optimization & Caching:** Redis caching layer untuk high-throughput, database indexing, horizontal scaling backend Go.
- **Legal & Compliance Finalization:** Audit kepatuhan UU PDP (UU No. 27 Tahun 2022) dan regulasi AML/CFT Bank Indonesia.
- **Public API & Webhooks (Beta):** Dokumentasi API publik untuk partner strategis (Phase 5 preparation).

---

## Success Criteria per Phase

| Phase | Milestone Kunci | Target Metrik |
|-------|----------------|---------------|
| **Phase 1** | Go-live MVP Barang di Fuji Testnet | 500 transaksi sukses, 0 critical bug |
| **Phase 2** | Peluncuran Jasa & Event Ticketing | 50 EO aktif, 1.000 tiket terjual |
| **Phase 3** | Vendor Marketplace live | 100 vendor terverifikasi onboarding |
| **Phase 4** | Mainnet migration & scaling | 100% audit log on C-Chain, MAU 5.000+ |
