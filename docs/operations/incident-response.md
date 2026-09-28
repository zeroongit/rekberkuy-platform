# Incident Response Plan — RekberKuy Platform

## 1. Overview
Prosedur tanggap darurat (*Incident Response*) mendefinisikan langkah-langkah deteksi, eskalasi, mitigasi, dan pemulihan ketika terjadi insiden keamanan, gangguan layanan, atau kegagalan infrastruktur pada platform RekberKuy.

## 2. Severities & Escalation

| Severity | Definisi | Waktu Tanggap | Eskalasi |
|----------|----------|---------------|----------|
| **Sev-1 (Critical)** | Kegagalan sistem total, kebocoran dana escrow, atau pelanggaran keamanan data masif. | < 15 menit | CTO, Founder, Lead DevOps |
| **Sev-2 (Major)** | Gangguan pada fitur utama (misal: pembayaran gagal, AI service down total). | < 1 jam | Engineering Lead |
| **Sev-3 (Moderate)**| Bug minor yang tidak menghentikan transaksi inti. | < 24 jam | On-call Engineer |

## 3. Response Steps
1. **Identifikasi & Containment:** Isolasi layanan yang terdampak (misalnya mengaktifkan maintenance mode atau memblokir IP mencurigakan).
2. **Mitigasi & Patching:** Perbaikan bug atau rollback ke versi stabil sebelumnya menggunakan feature flag / Docker image tag.
3. **Komunikasi:** Update status page (`status.rekberkuy.id`) dan pemberitahuan email kepada pengguna jika dana terdampak.
4. **Post-Mortem:** Analisis akar masalah (*Root Cause Analysis*) selambat-lambatnya 48 jam setelah insiden ditutup.
