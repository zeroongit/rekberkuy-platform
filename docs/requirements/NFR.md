# Non-Functional Requirements (NFR) — RekberKuy Platform

## 1. Overview
Dokumen ini menetapkan standar kualitas non-fungsional yang wajib dipenuhi oleh seluruh komponen arsitektur RekberKuy (Go backend, Next.js frontend, Python AI service, dan smart contract).

## 2. Rincian NFR

### 2.1 Performance & Scalability
- **API Response Time:** P50 < 100ms, P95 < 500ms, P99 < 1.000ms untuk seluruh endpoint core service.
- **Throughput:** Kapasitas normal 100 req/detik, dengan kemampuan spike hingga 1.000 req/detik pada peak event.
- **Frontend Load Time:** Largest Contentful Paint (LCP) < 2 detik pada dashboard web Next.js v16.
- **Horizontal Scaling:** Backend Go dirancang stateless, mendukung auto-scaling container di lingkungan Docker/Kubernetes.
- **Caching & Pooling:** Implementasi connection pooling PostgreSQL dan Redis caching untuk data referensi dan sesi.

### 2.2 Availability & Reliability
- **SLA Uptime:** Minimal 99.5% availability (maksimal downtime kumulatif ~43.8 jam per tahun).
- **RTO (Recovery Time Objective):** < 1 jam untuk pemulihan layanan setelah kegagalan infrastruktur.
- **RPO (Recovery Point Objective):** < 15 menit untuk titik pemulihan data transaksi.
- **Backup Strategy:** Database PostgreSQL dibackup secara otomatis setiap 6 jam dan disimpan selama 30 hari.

### 2.3 Security (Defense-in-Depth / ADR-0005)
- **Encryption:** Komunikasi wajib menggunakan TLS 1.3. Enkripsi data sensitif at-rest (KTP, nomor rekening) di database.
- **Session Security:** Token JWT disimpan dalam `HttpOnly`, `Secure`, dan `SameSite=Strict` cookies.
- **Input Validation:** Validasi ganda menggunakan Zod di frontend dan struct binding di backend Go.
- **SQL Injection Prevention:** Penggunaan parameterized queries di seluruh layer repository.
- **Rate Limiting:** Proteksi rate-limiting ketat pada endpoint otentikasi (max 10 req/s per IP).

### 2.4 Data Privacy & Compliance
- **UU PDP Compliance:** Kepatuhan penuh terhadap UU No. 27 Tahun 2022 (Perlindungan Data Pribadi). Hak akses, koreksi, dan penghapusan data pengguna didukung.
- **Data Retention:** Data keuangan dan transaksi disimpan minimal 5 tahun sesuai regulasi keuangan dan AML.
- **Data Residency:** Seluruh infrastruktur database utama ditempatkan di region server Indonesia.

### 2.5 Accessibility
- Mematuhi standar WCAG 2.1 Level AA untuk aksesibilitas antarmuka pengguna web.
- Kontras warna minimum 4.5:1 dan navigasi keyboard yang responsif.
