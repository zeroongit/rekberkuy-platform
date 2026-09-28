# Functional Requirements: Panel Admin (FR-ADMIN)

## 1. Overview
Modul Admin menyediakan kontrol penuh bagi tim internal RekberKuy untuk mengawasi transaksi, menyetujui KYC, memediasi sengketa, mengelola payout eksternal vendor, serta memantau metrik performa platform secara real-time.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-ADMIN-01** | Admin dapat memantau seluruh transaksi aktif maupun historis dalam sistem. | M | Implemented (`admin_handler.go`) |
| **FR-ADMIN-02** | Admin dapat melakukan review dan memberikan persetujuan (approve/reject) pada pengajuan KYC dan vendor. | M | Implemented |
| **FR-ADMIN-03** | Admin dapat mengambil alih sengketa, meninjau bukti, dan memutuskan outcome penyelesaian. | M | Implemented |
| **FR-ADMIN-04** | Admin dapat melakukan suspend atau ban pada akun pengguna yang terindikasi melakukan pelanggaran atau penipuan. | M | Implemented |
| **FR-ADMIN-05** | Admin memiliki akses ke dashboard metrik platform (Total GMV, MAU, Fraud Rate, Active Disputes). | M | Implemented |
| **FR-ADMIN-06** | Admin dapat menyesuaikan parameter konfigurasi sistem dan struktur fee platform. | S | Implemented |
