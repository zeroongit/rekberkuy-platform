# Functional Requirements: Transaksi Jasa / Milestone (FR-TRANSACTION-SERVICES)

## 1. Overview
Modul transaksi jasa mengatur alur kerja profesional dan freelancer berbasis tahapan (*milestone*), di mana seluruh nilai kontrak dikunci di muka namun dicairkan secara bertahap sesuai penyelesaian deliverable.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-TRX-SVC-01** | Klien dan penyedia jasa menyepakati kontrak jasa yang dibagi ke dalam beberapa milestone beserta nominalnya. | M | Implemented (`transaction_services_usecase.go`) |
| **FR-TRX-SVC-02** | Penyedia jasa menyetujui rincian kontrak sebelum pendanaan escrow dimulai. | M | Implemented |
| **FR-TRX-SVC-03** | Klien melakukan deposit nilai total kontrak ke rekening escrow. | M | Implemented |
| **FR-TRX-SVC-04** | Penyedia jasa menandai milestone selesai dan mengunggah bukti/file hasil kerja. | M | Implemented |
| **FR-TRX-SVC-05** | Klien mereview deliverable dan memberikan konfirmasi persetujuan milestone. | M | Implemented |
| **FR-TRX-SVC-06** | Sistem merilis dana tahap milestone terkait ke dompet penyedia jasa. | M | Implemented |
| **FR-TRX-SVC-07** | Auto-release berlaku pada milestone jika klien tidak merespons dalam batas waktu tertentu. | M | Implemented |
| **FR-TRX-SVC-08** | Penyedia jasa dapat merevisi pekerjaan berdasarkan catatan dari klien sebelum konfirmasi. | S | Implemented |
| **FR-TRX-SVC-09** | Pemotongan komisi platform dilakukan pada rilis milestone terakhir. | M | Implemented |
