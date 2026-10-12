# Use Case Specification per Aktor — RekberKuy

Dokumen ini merangkum daftar lengkap use case untuk setiap aktor yang berinteraksi dengan platform RekberKuy.

---

## 1. Buyer (Pembeli Barang / Klien Jasa / Pembeli Tiket)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-BUY-01** | Registrasi & Login Akun | Mendaftar dengan email dan password, verifikasi email, serta login ke dashboard. |
| **UC-BUY-02** | Top-up Saldo RekberPay | Mengisi saldo dompet via transfer bank (VA), QRIS, atau kartu kredit/debit (Midtrans). |
| **UC-BUY-03** | Membuat Transaksi Barang | Memilih penjual, memasukkan detail barang, dan mengunci dana ke escrow. |
| **UC-BUY-04** | Konfirmasi Penerimaan Barang | Memeriksa barang yang diterima dan mengkonfirmasi penerimaan untuk merilis dana ke penjual. |
| **UC-BUY-05** | Membuat Kontrak Jasa (Milestone) | Membuat kesepakatan jasa dengan penyedia, menentukan milestone dan nominal, serta lock dana. |
| **UC-BUY-06** | Konfirmasi Milestone Jasa | Mereview hasil kerja per milestone dan menyetujui rilis dana tahap tersebut. |
| **UC-BUY-07** | Membeli Tiket Event | Memilih kategori tiket, melakukan pembayaran, dan mendapatkan tiket digital ber-QR code. |
| **UC-BUY-08** | Membuka Sengketa (Dispute) | Melaporkan masalah (barang tidak sesuai/tidak sampai) dan mengunggah bukti untuk mediasi admin. |
| **UC-BUY-09** | Menarik Saldo (Withdraw) | Mencairkan saldo RekberPay ke rekening bank pribadi. |

---

## 2. Seller (Penjual Barang)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-SEL-01** | Verifikasi KYC Penjual | Mengunggah KTP dan selfie untuk verifikasi identitas wajib penjual. |
| **UC-SEL-02** | Menerima & Merespons Pesanan | Menerima notifikasi transaksi baru, mengonfirmasi kesanggupan, dan memasukkan nomor resi pengiriman. |
| **UC-SEL-03** | Menerima Pencairan Escrow | Menerima dana hasil penjualan ke dompet RekberPay setelah pembeli konfirmasi atau auto-release. |
| **UC-SEL-04** | Merespons Sengketa | Memberikan argumen dan bukti tanding saat pembeli membuka sengketa transaksi. |

---

## 3. Service Provider (Penyedia Jasa / Freelancer)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-SVC-01** | Menerima Kontrak Jasa | Menerima tawaran kontrak jasa dengan rincian milestone dari klien. |
| **UC-SVC-02** | Submit Deliverable Milestone | Menandai milestone selesai dan mengunggah tautan/file hasil kerja ke klien. |
| **UC-SVC-03** | Meminta Revisi / Pencairan | Meminta persetujuan milestone atau melakukan penyesuaian berdasarkan feedback klien. |
| **UC-SVC-04** | Menerima Pembayaran Bertahap | Menerima rilis dana per milestone ke dompet setelah disetujui klien. |

---

## 4. Event Organizer (EO)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-EO-01** | Pendaftaran & KYC EO | Melakukan registrasi dan verifikasi KYC khusus Event Organizer. |
| **UC-EO-02** | Membuat & Publish Event | Mengatur detail event (nama, tanggal, lokasi, kapasitas, harga tiket) dan mempublikasikannya. |
| **UC-EO-03** | Membuka Kontrak Vendor | Mencari vendor di marketplace platform atau mendaftarkan vendor eksternal, serta membuat kontrak escrow. |
| **UC-EO-04** | Mengelola Payout Vendor | Menandai pekerjaan vendor selesai dan memicu rilis pembayaran (internal wallet atau external disbursement). |
| **UC-EO-05** | Audit & Surplus Event | Menyelesaikan event, menjalankan kalkulator keuangan event, menerima bonus EO, dan memicu refund surplus peserta. |

---

## 5. Vendor (Vendor Pendukung Event)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-VEN-01** | Pendaftaran Profil Vendor | Mendaftar sebagai vendor (katering, dekorasi, dll.) dan upload legalitas (SIUP/NPWP). |
| **UC-VEN-02** | Manajemen Katalog & Portofolio | Mengatur layanan, paket harga, dan portofolio agar tampil di vendor marketplace. |
| **UC-VEN-03** | Menerima Kontrak & Milestone | Menerima kontrak dari EO, mengerjakan event sesuai jadwal, dan menerima pembayaran escrow. |

---

## 6. Admin (Tim Internal RekberKuy)

| Use Case ID | Nama Use Case | Deskripsi |
|-------------|---------------|-----------|
| **UC-ADM-01** | Review & Approve KYC | Memeriksa dokumen KYC pengguna/vendor dan memberikan status approved/rejected. |
| **UC-ADM-02** | Mediasi Sengketa (Dispute) | Membaca bukti dari kedua belah pihak dan memutuskan outcome (`REFUND_BUYER` / `RELEASE_TO_SELLER`). |
| **UC-ADM-03** | Monitoring Platform Metrics | Mengamati dashboard GMV, MAU, fraud rate, dan status sistem secara real-time. |
| **UC-ADM-04** | Manually Trigger Disbursement | Memeriksa dan menandai payout vendor eksternal yang dibayar via Midtrans disbursement (`DISBURSED`). |
| **UC-ADM-05** | Konfigurasi Platform & Fee | Mengatur parameter platform, threshold fraud, dan struktur fee jika diperlukan. |
