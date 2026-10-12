# Functional Requirements: CRM Loyalty Tiering (FR-CRM)

## 1. Overview
Modul CRM Loyalty mengevaluasi volume transaksi bulanan pengguna untuk menempatkan mereka ke dalam tier loyalitas (Bronze, Silver, Gold, Platinum) yang memberikan keuntungan berupa potongan komisi fee platform dan prioritas layanan.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-CRM-01** | Sistem menjalankan evaluasi tier pengguna secara otomatis setiap tanggal 1 setiap bulan melalui background worker (`crm_worker.go`). | M | Implemented (`crm_worker.go`) |
| **FR-CRM-02** | Terdapat 4 level tier loyalitas: Bronze, Silver, Gold, dan Platinum berdasarkan GMV bulanan. | M | Implemented (`profile.go`) |
| **FR-CRM-03** | Pengguna di tier lebih tinggi menikmati insentif seperti diskon take rate platform dan prioritas penanganan mediasi. | S | Implemented (`finance_calculator.go`) |
| **FR-CRM-04** | Pengguna menerima notifikasi ketika tier loyalitas mereka naik atau turun. | S | Implemented |
