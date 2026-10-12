# Threat Model — RekberKuy Platform

## 1. Overview
Threat model ini mengidentifikasi potensi ancaman terhadap platform escrow RekberKuy serta mitigasi keamanan yang diterapkan pada setiap lapisan sistem.

## 2. Threat Analysis (STRIDE Matrix)

| Threat Category | Potential Attack Vector | Mitigation / Countermeasure |
|-----------------|------------------------|-----------------------------|
| **Spoofing** | Penyerang mencuri token sesi atau memalsukan identitas pengguna. | Penyimpanan token dalam `HttpOnly` cookies, verifikasi JWT ketat pada middleware. |
| **Tampering** | Memanipulasi parameter jumlah uang atau status transaksi via client request. | Validasi Zod & Go struct binding, validasi state machine di Usecase layer. |
| **Repudiation** | Pengguna menyangkal telah melakukan konfirmasi rilis dana atau transaksi. | Audit trail immutable lengkap di database dan on-chain audit log di Avalanche (`TransactionLogger.sol`). |
| **Information Disclosure** | Kebocoran data sensitif (KTP, nomor rekening, password). | Enkripsi at-rest untuk data sensitif di PostgreSQL, hashing password dengan bcrypt cost 12. |
| **Denial of Service (DoS)** | Banjir request ke endpoint auth atau transaksi untuk melumpuhkan server. | Rate limiting ketat per IP (max 10 req/s untuk auth), Nginx reverse proxy protection. |
| **Elevation of Privilege** | Pengguna biasa mengakses endpoint Admin atau melakukan aksi mediasi. | Role-Based Access Control (RBAC) middleware yang memeriksa claim role secara eksplisit. |
