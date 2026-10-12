# Functional Requirements: Dompet RekberPay & Ledger (FR-WALLET)

## 1. Overview
Modul dompet mengelola saldo pengguna (RekberPay), riwayat mutasi keuangan (ledger), integrasi top-up dengan Midtrans, serta pencairan dana (withdrawal) dengan jaminan konsistensi transaksi ACID dan idempotensi.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-WALLET-01** | Pengguna dapat melihat saldo RekberPay (saldo tersedia dan saldo dalam escrow) secara real-time. | M | Implemented (`wallet_usecase.go`) |
| **FR-WALLET-02** | Pengguna dapat melakukan top-up saldo melalui transfer bank (Virtual Account), QRIS, atau kartu pembayaran via Midtrans Snap. | M | Implemented (`midtrans/client.go`) |
| **FR-WALLET-03** | Pengguna dapat menarik saldo (withdraw) ke rekening bank terverifikasi atas nama yang sama. | M | Implemented |
| **FR-WALLET-04** | Sistem melarang saldo menjadi negatif dalam kondisi apapun melalui database check constraint dan locking. | M | Implemented |
| **FR-WALLET-05** | Setiap mutasi saldo wajib dijalankan dalam transaksi database ACID menggunakan `SELECT FOR UPDATE`. | M | Implemented (`unit_of_work.go`) |
| **FR-WALLET-06** | Pengguna dapat melihat riwayat lengkap seluruh mutasi saldo (ledger) dengan rincian status transaksi. | M | Implemented |
| **FR-WALLET-07** | Sistem menggunakan Idempotency Key pada setiap mutasi finansial untuk mencegah duplikasi debet/kredit. | M | Implemented (`idempotency_middleware.go`) |
| **FR-WALLET-08** | Dana yang tertahan dalam escrow ditampilkan secara terpisah dari saldo bebas. | M | Implemented |
| **FR-WALLET-09** | Batasan limit top-up dan withdraw harian mematuhi regulasi Bank Indonesia. | M | Implemented |
| **FR-WALLET-10** | Notifikasi dikirimkan setiap kali terjadi mutasi saldo yang sukses. | S | Implemented |
