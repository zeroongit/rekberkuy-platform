# Spesifikasi Arsitektur Antarmuka, Routing, dan Desain Beranda — RekberKuy Platform

Dokumen ini merinci arsitektur routing antarmuka frontend (Next.js v16 App Router), aturan akses (RBAC & Account Governance), serta spesifikasi fungsional halaman Beranda (Discovery Hub) yang mengadopsi pola penemuan produk/layanan ala e-commerce modern (seperti Shopee).

---

## 1. Pemisahan Struktur Halaman & Routing Architecture

RekberKuy memisahkan struktur antarmuka menjadi tiga zona utama berdasarkan status autentikasi dan peran pengguna (`RBAC`):

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Front-End Architecture                          │
└───────┬────────────────────────┬───────────────────────────┬───────────┘
        │                        │                           │
        ▼                        ▼                           ▼
┌──────────────┐         ┌───────────────┐           ┌───────────────┐
│ Landing Page │         │ Beranda (Hub) │           │   Dashboard   │
│    (Public)  │         │  (Auth User)  │           │  (RBAC Gated) │
└──────────────┘         └───────────────┘           └───────────────┘
```

### A. Landing Page (`/`)
- **Tujuan:** Pintu masuk publik (public entry point) untuk pengunjung dan calon pengguna baru.
- **Isi & Komponen:**
  - Hero Section: Penjelasan nilai inti platform (escrow aman untuk Barang, Jasa, dan Event dengan audit on-chain Avalanche gasless).
  - Feature Highlights: Keunggulan perlindungan pembeli & penjual, verifikasi KYC AI, dan garansi escrow.
  - Fee Transparency Table: Tabel rincian komisi platform (Goods 1.5%, Services 2.0%, Events 2.0%, Vendor 1.5%).
  - Call to Action (CTA): Tombol Autentikasi (`Login` / `Register`).
- **Akses:** Publik (tidak memerlukan token autentikasi).

### B. Beranda / Discovery Hub (`/beranda`)
- **Tujuan:** Halaman utama setelah login khusus untuk Akun Personal (`user`) sebagai pusat eksplorasi dan penemuan produk, jasa, dan event.
- **Konsep UX/UI:** Mengadopsi pola e-commerce modern (ala Shopee) dengan fokus pada kecepatan penemuan (discovery), banner promosi, navigasi kategori bertingkat, dan modul rekomendasi dinamis.
- **Akses:** Pengguna yang telah terautentikasi dengan role `user`.

### C. Dashboard Privat (`/dashboard/*` & `/admin`)
- **Tujuan:** Area operasional privat berpagar RBAC yang disesuaikan per role bisnis dan administratif (`user`, `seller`, `service`, `eo`, `vendor`, `admin`).
- **Struktur Routing Dashboard:**
  - `/dashboard/user`: Manajemen transaksi barang, riwayat kontrak jasa, tiket event QR code, dan saldo RekberPay.
  - `/dashboard/seller`: Manajemen pesanan masuk, resi pengiriman, KYC status, dan pencairan escrow barang.
  - `/dashboard/service`: Manajemen milestone kontrak jasa, upload deliverable, dan pencairan bertahap.
  - `/dashboard/eo`: Pembuatan & pengelolaan event, sistem tiket, kontrak vendor, dan kalkulator keuangan event.
  - `/dashboard/vendor`: Manajemen katalog layanan vendor, portofolio, dan penerimaan kontrak escrow EO.
  - `/admin`: Panel kontrol khusus Admin RekberKuy untuk review KYC, mediasi sengketa (`DISPUTE`), monitoring metrik GMV/MAU, dan audit relayer.

---

## 2. Spesifikasi Fungsional Halaman Beranda (Discovery Hub)

Halaman Beranda dirancang sebagai pusat eksplorasi interaktif yang terhubung langsung dengan backend `core-service`. Berikut adalah spesifikasi fungsional utama untuk panduan tim Frontend Next.js:

### A. Global Search Bar (Pencarian Instan)
- **Letak:** Bagian atas halaman Beranda (sticky header).
- **Fungsi:** Bar pencarian universal untuk menemukan barang, jasa, vendor, atau event secara instan.
- **Behavior & API Integration:**
  - Mendukung auto-suggestion saat mengetik (debounce 300ms).
  - Query parameter dikirim ke backend untuk memfilter entitas lintas domain (Barang, Jasa, Event, Vendor).
  - Navigasi hasil pencarian ke halaman katalog/direktori dengan filter aktif.

### B. 3-Tier Category Taxonomy Navigation (Navigasi Kategori Bertingkat)
- **Sumber Data:** Merujuk pada data kategori 3-tier di backend `core-service` (`GoodsCategory`, `ServiceCategory`, `EventCategory`, `VendorSubCategory` beserta sub-kategori dan sub-sub-kategori).
- **Struktur UI:**
  - **Tier 1 (Main Domain/Category):** Tab atau Grid utama (Goods, Services, Events, Vendors).
  - **Tier 2 (Sub-Category):** Daftar kategori turunan (misal: Elektronik, Fashion untuk Goods; Desain Grafis, Web Development untuk Services; Seminar, Musik untuk Events).
  - **Tier 3 (Sub-Sub-Category / Specific Specialty):** Spesifikasi detail item.
- **Interaksi:** Mengklik kategori akan memfilter section produk/layanan di bawahnya secara dinamis tanpa reload halaman (Client-side fetching via React Server Components / SWR / Zustand).

### C. Best Seller & Trending Section (Modul Dinamis Berbasis Database)
- **Tujuan:** Menampilkan modul produk terlaris, event mendatang, atau vendor pilihan berdasarkan kategori yang sedang dilihat.
- **Modul Utama:**
  1. **Flash Deals / Trending This Week:** Produk barang fisik/digital dengan volume transaksi tertinggi dalam 7 hari terakhir.
  2. **Featured Events & Upcoming Concerts:** Event pilihan dengan sisa kuota tiket terbaca secara real-time dari escrow event.
  3. **Top Rated Vendors:** Daftar vendor event (katering, dekorasi, sound system) dengan badge "Terverifikasi" dan rating tertinggi.
- **Implementasi Frontend:**
  - Komponen modular (Card Component) yang reusable dengan TypeScript strict interface.
  - Mengambil data melalui API endpoints `core-service` dengan caching Redis di backend untuk performa tinggi (P95 < 500ms).

---

## 3. Aturan Akses & Governance (RBAC & Account Separation)

Untuk menjaga integritas platform dan kepatuhan hukum, diberlakukan pemisahan tegas:
- **Akun Personal (`user`):** Diizinkan mengakses Beranda (`/beranda`) dan dashboard user (`/dashboard/user`) untuk berbelanja, memesan jasa, dan membeli tiket event.
- **Akun Komersial (`seller`, `service`, `eo`, `vendor`):** Memiliki dashboard operasional mandiri masing-masing (`/dashboard/seller`, `/dashboard/service`, `/dashboard/eo`, `/dashboard/vendor`) dan **dilarang keras membeli tiket event** (di-enforce oleh middleware `RequirePersonalAccount()` dengan respons HTTP 403 Forbidden).
- **Admin (`admin`):** Akses penuh ke `/admin` untuk moderasi KYC dan mediasi sengketa.
