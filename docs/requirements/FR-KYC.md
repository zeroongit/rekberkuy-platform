# Functional Requirements: Verifikasi Identitas & KYC (FR-KYC)

## 1. Overview
Modul KYC (Know Your Customer) memastikan bahwa seluruh Penjual, Event Organizer, dan Vendor telah diverifikasi identitasnya untuk mencegah penipuan dan mematuhi regulasi Anti-Pencucian Uang (AML). Arsitektur mengikuti prinsip **"AI scores, admin decides"**: AI service (`backend-ai`) memberikan skor keyakinan, namun keputusan akhir mutlak di tangan Admin.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-KYC-01** | Penjual, EO, dan Vendor wajib mengajukan KYC sebelum dapat bertransaksi sebagai pihak penerima dana. | M | Implemented (`kyc_usecase.go`) |
| **FR-KYC-02** | Pengguna mengunggah foto KTP (depan & belakang) serta foto selfie memegang KTP. | M | Implemented |
| **FR-KYC-03** | Sistem AI (`backend-ai` via Groq) melakukan pre-screening otomatis terhadap dokumen dan mengembalikan skor keyakinan (`ai_score` & `ai_reason`). | M | Implemented (`kyc/client.go`) |
| **FR-KYC-04** | Admin mereview pengajuan KYC melalui panel admin dan memberikan keputusan Approve/Reject dalam 1x24 jam. | M | Implemented (`admin_handler.go`) |
| **FR-KYC-05** | Pengguna menerima notifikasi email mengenai hasil verifikasi KYC mereka. | M | Implemented |
| **FR-KYC-06** | Pengguna yang ditolak dapat mengajukan ulang KYC dengan dokumen yang diperbarui. | M | Implemented |
| **FR-KYC-07** | Akun yang lolos verifikasi mendapatkan badge "Terverifikasi" dan promosi role otomatis dalam unit of work yang sama. | M | Implemented |
| **FR-KYC-08** | Pengecekan OCR otomatis tanpa intervensi manual untuk dokumen dengan kualitas tinggi. | C | Planned |
