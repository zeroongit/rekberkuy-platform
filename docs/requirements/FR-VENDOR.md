# Functional Requirements: Vendor Marketplace (FR-VENDOR)

## 1. Overview
Modul vendor marketplace menyediakan wadah bagi Event Organizer untuk menemukan, memfilter, dan mengontrak vendor pendukung event (katering, dekorasi, fotografer, sound system) dengan jaminan pembayaran escrow.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-VENDOR-01** | Penyedia jasa event dapat mendaftar sebagai Vendor dengan melengkapi profil bisnis dan kategori layanan. | M | Implemented (`vendor_usecase.go`) |
| **FR-VENDOR-02** | Vendor wajib mengunggah dokumen legalitas usaha (SIUP, NPWP, portofolio). | M | Implemented |
| **FR-VENDOR-03** | Admin memverifikasi profil dan legalitas vendor dalam waktu maksimal 2x24 jam. | M | Implemented |
| **FR-VENDOR-04** | Vendor yang telah disetujui tampil di direktori marketplace dengan badge terverifikasi. | M | Implemented |
| **FR-VENDOR-05** | Event Organizer dapat mencari dan memfilter vendor berdasarkan kategori, kota, dan rentang harga. | M | Implemented |
| **FR-VENDOR-06** | EO dapat melihat profil detail, portofolio, dan ulasan/rating dari vendor. | M | Implemented |
| **FR-VENDOR-07** | Fasilitas komunikasi/chat antara EO dan vendor terintegrasi dalam platform. | S | Implemented (Basic) |
