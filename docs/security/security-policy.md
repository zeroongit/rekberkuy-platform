# Security Policy — RekberKuy Platform

## 1. Overview
Kebijakan keamanan ini mendefinisikan standar perlindungan data, manajemen kerentanan, enkripsi, dan tata kelola akses untuk seluruh infrastruktur dan layanan RekberKuy.

## 2. Security Principles (Defense-in-Depth / ADR-0005)
1. **Zero Trust Network:** Semua komunikasi internal dan eksternal dienkripsi menggunakan TLS 1.3. Layanan database dan cache terisolasi dalam jaringan privat Docker/VPS.
2. **Secure Session Management:** Token autentikasi JWT disimpan secara eksklusif di dalam `HttpOnly`, `Secure`, dan `SameSite=Strict` cookies.
3. **Strict Validation:** Validasi input berlapis menggunakan Zod di frontend dan struct binding ketat di backend Go untuk mencegah injection attack.
4. **Idempotency & ACID:** Seluruh operasi finansial mewajibkan Idempotency Key dan transaksi database ACID dengan `SELECT FOR UPDATE`.

## 3. Vulnerability Disclosure & Incident Reporting
Jika Anda menemukan celah keamanan atau kerentanan pada platform RekberKuy, harap segera laporkan melalui email ke `security@rekberkuy.id`. Jangan mempublikasikan kerentanan sebelum tim security kami memberikan konfirmasi perbaikan.
