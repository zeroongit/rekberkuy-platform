# Functional Requirements: Transaksi Event & Tiket (FR-TRANSACTION-EVENTS)

## 1. Overview
Modul transaksi event mengelola pengadaan event oleh Event Organizer (EO), penjualan tiket kepada peserta dengan tiket QR unik, pembuatan kontrak vendor dengan pembayaran escrow, hingga perhitungan audit keuangan event dan distribusi surplus.

**Aturan Akses Pembelian Tiket:**
Pembelian tiket event (`POST /api/v1/events/:id/tickets` atau `/transactions/events/lock`) dibatasi secara ketat hanya untuk **Akun Personal (USER)**. Akun Komersial (`EVENT_ORGANIZER`, `VENDOR`, `SELLER`, `SERVICE_PROVIDER`) dilarang keras membeli tiket event dan akan ditolak otomatis oleh sistem dengan respon HTTP 403 Forbidden.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-TRX-EVT-01** | EO dapat membuat event baru dengan informasi lengkap (nama, tanggal, lokasi, kapasitas, dan harga tiket). | M | Implemented (`transaction_events_usecase.go`) |
| **FR-TRX-EVT-02** | EO mengelola kategori tiket (VIP, Regular, Presale). | S | Implemented |
| **FR-TRX-EVT-03** | Peserta (Akun Personal / `USER`) dapat membeli tiket event melalui platform. Akun komersial dilarang keras membeli tiket. | M | Implemented (`auth_middleware.go` + `RequirePersonalAccount`) |
| **FR-TRX-EVT-04** | Sistem menghasilkan tiket digital dengan QR Code unik untuk check-in. | M | Implemented |
| **FR-TRX-EVT-05** | Dana penjualan tiket ditampung dalam escrow event. | M | Implemented |
| **FR-TRX-EVT-06** | EO dapat membuat kontrak vendor pendukung dari dalam panel event. | M | Implemented |
| **FR-TRX-EVT-07** | EO dapat memilih vendor dari marketplace internal atau mendaftarkan vendor eksternal. | M | Implemented |
| **FR-TRX-EVT-08** | Pembayaran vendor menggunakan escrow milestone. | M | Implemented |
| **FR-TRX-EVT-09** | Setelah event selesai, sistem menjalankan `CalculateEventAudit` untuk memotong fee platform 5% dan mendistribusikan bonus EO. | M | Implemented (`finance_calculator.go`) |
| **FR-TRX-EVT-10** | Jika event dibatalkan, sistem otomatis melakukan refund penuh ke dompet peserta. | M | Implemented |
| **FR-TRX-EVT-11** | Kontrak vendor dikenakan fee platform 1.5%. | M | Implemented |
