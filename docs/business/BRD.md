# Business Requirements Document (BRD)
# RekberKuy Platform

| | |
|---|---|
| **Dokumen** | Business Requirements Document |
| **Status** | Draft v1.0 |
| **Versi** | 1.0.0 |
| **Tanggal** | September 2026 |
| **Author** | Karis (Founder & Product Owner) |
| **Reviewer** | Business Stakeholders, Legal, Finance |
| **Klasifikasi** | Confidential — Internal Only |

---

## Version History

| Versi | Tanggal | Author | Perubahan |
|-------|---------|--------|-----------|
| 0.1 | Agt 2026 | Karis | Draft awal konsep bisnis |
| 0.2 | Sep 2026 | Karis | Tambah modul Event & Vendor Marketplace |
| 1.0 | Sep 2026 | Karis | BRD lengkap sesuai README platform terkini |

---

## Daftar Isi

1. [Executive Summary](#1-executive-summary)
2. [Latar Belakang & Konteks Bisnis](#2-latar-belakang--konteks-bisnis)
3. [Tujuan Bisnis](#3-tujuan-bisnis)
4. [Analisis Current State (As-Is)](#4-analisis-current-state-as-is)
5. [Desired State (To-Be)](#5-desired-state-to-be)
6. [Stakeholder & Kebutuhan](#6-stakeholder--kebutuhan)
7. [Business Requirements](#7-business-requirements)
8. [Business Rules](#8-business-rules)
9. [Asumsi & Batasan](#9-asumsi--batasan)
10. [Dependensi Bisnis](#10-dependensi-bisnis)
11. [Risiko Bisnis](#11-risiko-bisnis)
12. [Kriteria Keberhasilan](#12-kriteria-keberhasilan)
13. [Glossary](#13-glossary)

---

## 1. Executive Summary

**RekberKuy** adalah platform escrow dan penyelesaian pembayaran (*payment settlement*) yang dirancang khusus untuk mengamankan transaksi lintas tiga domain utama: **Barang** (fisik dan digital), **Jasa** (profesional dan freelance), dan **Event** (pengadaan, vendor, dan tiket).

Platform ini memposisikan diri sebagai infrastruktur kepercayaan digital bagi pelaku ekonomi Indonesia — dari reseller individu, freelancer, hingga Event Organizer skala menengah — yang selama ini tidak memiliki akses ke mekanisme escrow formal yang mudah, cepat, dan terjangkau.

**Proposisi nilai inti:**
- Dana pembeli dikunci di sistem escrow dan hanya dilepaskan saat kondisi bisnis terpenuhi
- Setiap transaksi yang selesai dicatat secara permanen di blockchain Avalanche sebagai audit log yang tidak dapat dimanipulasi — tanpa membebani pengguna dengan gas fee atau keharusan memiliki wallet kripto
- Satu platform untuk tiga domain — Barang, Jasa, dan Event — dengan manajemen vendor terintegrasi

**Tagline:** *Tailored Escrow & Payment Settlement System for Goods, Services, and Events*

---

## 2. Latar Belakang & Konteks Bisnis

### 2.1 Lanskap Masalah

Indonesia adalah salah satu ekonomi digital terbesar di Asia Tenggara, namun infrastruktur kepercayaan antar pihak yang bertransaksi masih sangat lemah di luar ekosistem marketplace besar. Mayoritas transaksi peer-to-peer masih dilakukan melalui transfer bank langsung — tanpa jaminan, tanpa mediasi, dan tanpa bukti yang tidak bisa dipalsukan.

Tiga segmen yang paling terdampak:

**Segmen Barang:** Transaksi jual beli di luar marketplace besar (WhatsApp, Instagram, komunitas) tidak memiliki perlindungan. Fraud marak — barang tidak dikirim setelah dibayar, atau pembeli klaim tidak terima barang.

**Segmen Jasa:** Tidak ada mekanisme formal yang mudah diakses untuk memastikan pembayaran jasa terlaksana sesuai progres. Freelancer sering tidak dibayar; klien sering kecewa karena pekerjaan tidak sesuai ekspektasi.

**Segmen Event:** Ekosistem event masih sangat manual. EO mengelola pembayaran ke puluhan vendor lewat transfer manual tanpa jejak audit yang rapi. Vendor tidak punya jaminan pembayaran. Tidak ada platform yang menghubungkan EO dengan vendor terverifikasi secara terintegrasi.

### 2.2 Peluang Bisnis

- Pasar rekber informal di Indonesia diestimasi bernilai triliunan rupiah per tahun, sebagian besar tidak terlindungi
- Belum ada pemain yang melayani tiga domain (Barang + Jasa + Event) dalam satu ekosistem terintegrasi
- Segmen vendor event adalah blue ocean — belum ada marketplace vendor event yang terstruktur di Indonesia
- Adopsi teknologi blockchain untuk transparansi semakin diterima, terutama pasca beberapa kasus fraud besar

### 2.3 Positioning

RekberKuy bukan marketplace — RekberKuy adalah **lapisan kepercayaan** yang bisa digunakan oleh siapapun yang bertransaksi, tanpa harus berpindah dari platform komunikasi yang sudah mereka pakai.

---

## 3. Tujuan Bisnis

### 3.1 Tujuan Jangka Pendek (0–6 Bulan)

| ID | Tujuan | Metrik Sukses |
|----|--------|---------------|
| BO-01 | Membuktikan product-market fit di segmen Barang | 500 transaksi barang berhasil diselesaikan |
| BO-02 | Membangun kepercayaan pengguna awal | NPS > 40 dari pengguna beta |
| BO-03 | Memvalidasi model monetisasi fee transaksi | Revenue dari fee > Rp 10 Juta di bulan ke-3 |
| BO-04 | Memastikan keamanan sistem sebelum scale | Zero fraud incident di fase beta |

### 3.2 Tujuan Jangka Menengah (6–18 Bulan)

| ID | Tujuan | Metrik Sukses |
|----|--------|---------------|
| BO-05 | Ekspansi ke segmen Jasa dan Event | 50 EO aktif, 100 vendor terverifikasi |
| BO-06 | Membangun ekosistem vendor event terbesar di Indonesia | 500 vendor terdaftar di marketplace |
| BO-07 | Mencapai volume transaksi yang sustain | GMV Rp 2 Miliar/bulan |
| BO-08 | Produksi audit log on-chain | 100% transaksi selesai tercatat di Avalanche Mainnet |

### 3.3 Tujuan Jangka Panjang (18 Bulan+)

| ID | Tujuan |
|----|--------|
| BO-09 | Menjadi platform escrow paling dipercaya di Indonesia |
| BO-10 | Ekspansi ke mobile app |
| BO-11 | Open API untuk integrasi pihak ketiga (white-label escrow) |
| BO-12 | Program kredit berbasis riwayat transaksi RekberKuy |

---

## 4. Analisis Current State (As-Is)

### 4.1 Cara Bertransaksi Saat Ini (Tanpa RekberKuy)

**Transaksi Barang:**
```
Pembeli ──[transfer bank]──► Penjual
         (tidak ada jaminan)
```
Risiko: pembeli transfer tapi barang tidak dikirim, atau barang dikirim tapi pembeli klaim tidak terima.

**Transaksi Jasa:**
```
Klien ──[DP 50% transfer]──► Freelancer ──[kerja]──► Klien
      ──[Pelunasan?]──► Sering bermasalah di tahap ini
```
Risiko: freelancer tidak menyelesaikan pekerjaan setelah terima DP, atau klien menolak bayar pelunasan tanpa alasan jelas.

**Transaksi Event:**
```
EO ──[transfer ke vendor 1]──► Vendor 1
   ──[transfer ke vendor 2]──► Vendor 2
   ──[transfer ke vendor N]──► Vendor N
   (manual, tidak ada jejak audit, tidak ada jaminan)
```
Risiko: vendor tidak deliver setelah terima DP, EO tidak punya leverage, tidak ada satu dashboard yang konsolidasikan semua pengeluaran event.

### 4.2 Pain Point per Segmen

| Segmen | Pain Point | Dampak Bisnis |
|--------|-----------|---------------|
| Pembeli Barang | Tidak ada jaminan barang dikirim | Kehilangan uang, trauma transaksi online |
| Penjual Barang | Tidak ada bukti transaksi yang tidak bisa dipalsukan | Rentan dispute palsu |
| Freelancer | Tidak ada jaminan pembayaran bertahap | Sering tidak dibayar |
| Klien Jasa | Tidak ada jaminan pekerjaan selesai | Uang hilang, pekerjaan tidak selesai |
| Event Organizer | Tidak ada dashboard konsolidasi vendor | Pengeluaran tidak terkontrol |
| Vendor Event | Tidak ada jaminan dibayar setelah event | Cash flow terganggu |

---

## 5. Desired State (To-Be)

### 5.1 Alur Bisnis Target

**Transaksi Barang:**
```
Pembeli ──[kunci dana]──► Escrow RekberKuy ──[konfirmasi terima]──► Penjual
        ←─[jaminan]──────────────────────── ←─[mediasi jika sengketa]
```

**Transaksi Jasa (Milestone):**
```
Klien ──[kunci total kontrak]──► Escrow
      ←──[milestone 1 selesai, approve]──► Release Milestone 1 ──► Freelancer
      ←──[milestone 2 selesai, approve]──► Release Milestone 2 ──► Freelancer
      ... dst
```

**Transaksi Event:**
```
Peserta ──[beli tiket]──► Escrow Event
EO ──[alokasi vendor]──► Kontrak Vendor ──[invoice + approve]──► Payout Vendor
EO ──[event selesai]──► Release pendapatan tiket ke EO (dikurangi fee)
```

### 5.2 Ekosistem Platform

```
┌─────────────────────────────────────────────────────┐
│                  RekberKuy Platform                  │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │  Escrow  │  │  Wallet  │  │  Vendor           │  │
│  │  Barang  │  │RekberPay │  │  Marketplace      │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │  Escrow  │  │  KYC &   │  │  Blockchain       │  │
│  │  Jasa    │  │  Fraud   │  │  Audit Log        │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │  Escrow  │  │  CRM     │  │  Admin &          │  │
│  │  Event   │  │  Loyalty │  │  Dispute          │  │
│  └──────────┘  └──────────┘  └───────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 6. Stakeholder & Kebutuhan

### 6.1 Internal Stakeholder

| Stakeholder | Peran | Kebutuhan Utama |
|-------------|-------|-----------------|
| Founder (Karis) | Product Owner, keputusan strategis | Platform tumbuh, sustainable revenue, compliant |
| Admin RekberKuy | Operasional harian, mediasi, KYC review | Tool yang efisien untuk moderasi & monitoring |
| Tim Finance (future) | Rekonsiliasi, laporan keuangan | Laporan transaksi akurat, audit trail lengkap |
| Tim Legal (future) | Compliance, ToS, kontrak | Dokumentasi legal lengkap, data retention sesuai regulasi |

### 6.2 External Stakeholder (Pengguna Platform)

#### Pengguna (User / Pembeli)
- **Kebutuhan:** Bisa membeli dengan aman, dana terlindungi sampai barang/jasa diterima
- **Ekspektasi:** Alur mudah (< 5 langkah), notifikasi real-time, status transaksi transparan
- **Concern:** Dana tidak hilang jika terjadi masalah

#### Verified Merchant (Penjual Terverifikasi)
- **Kebutuhan:** Kepastian pembayaran, reputasi terbangun, mudah kelola pesanan
- **Ekspektasi:** Dana cair cepat setelah transaksi selesai, dashboard penjualan yang informatif
- **Concern:** Dispute palsu dari pembeli, waktu pencairan dana

#### Event Organizer (EO)
- **Kebutuhan:** Satu dashboard untuk kelola seluruh keuangan event, akses ke vendor terverifikasi
- **Ekspektasi:** Bisa track semua pengeluaran vendor real-time, tiket terjual terekam otomatis
- **Concern:** Vendor tidak deliver setelah dibayar DP

#### Vendor
- **Kebutuhan:** Jaminan pembayaran, akses ke klien EO baru, proses klaim pembayaran yang mudah
- **Ekspektasi:** Payout setelah upload invoice dan disetujui EO, profil bisnis tampil di marketplace
- **Concern:** EO tidak bayar setelah event selesai

#### Payment Gateway (Midtrans)
- **Kebutuhan:** Integrasi teknis yang sesuai standar, idempotency, webhook reliable
- **Ekspektasi:** Platform comply dengan syarat dan ketentuan Midtrans

#### Regulator (Bank Indonesia, OJK)
- **Kebutuhan:** Platform beroperasi sesuai regulasi, ada audit trail, KYC dijalankan
- **Ekspektasi:** Transparansi, laporan jika diperlukan, data residency di Indonesia

---

## 7. Business Requirements

> Format: `[ID] Kebutuhan Bisnis` — Prioritas: **M** (Must) / **S** (Should) / **C** (Could) / **W** (Won't)
> 
> BRD mendefinisikan **apa yang dibutuhkan bisnis** — bukan bagaimana sistem mengimplementasikannya.

---

### BR-01: Manajemen Identitas & Akses Berbasis Peran

Platform harus mendukung lima peran pengguna dengan hak akses yang berbeda:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-01.1 | Platform membedakan akses berdasarkan peran: **User** (pembeli umum), **Verified Merchant** (penjual terverifikasi), **Vendor** (penyedia layanan event), **Event Organizer**, dan **Admin** | M |
| BR-01.2 | Pengguna baru masuk sebagai User dan dapat mengajukan upgrade peran ke Verified Merchant, Vendor, atau EO | M |
| BR-01.3 | Upgrade peran memerlukan verifikasi identitas (KYC) yang disetujui Admin | M |
| BR-01.4 | Admin memiliki akses penuh ke seluruh data platform untuk keperluan monitoring dan moderasi | M |
| BR-01.5 | Setiap peran hanya dapat mengakses modul yang relevan dengan aktivitasnya | M |

---

### BR-02: Escrow & Penyelesaian Pembayaran — Barang

Platform harus menyediakan mekanisme escrow yang aman untuk transaksi barang fisik dan digital:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-02.1 | Dana pembeli dikunci di sistem escrow saat transaksi dibuat — tidak dapat diakses penjual sampai kondisi terpenuhi | M |
| BR-02.2 | Penjual harus mengkonfirmasi pengiriman dan memasukkan nomor resi sebelum dana dilepaskan | M |
| BR-02.3 | Pembeli mengkonfirmasi penerimaan barang untuk memicu pelepasan dana ke penjual | M |
| BR-02.4 | Jika pembeli tidak mengkonfirmasi dalam batas waktu yang ditetapkan, sistem secara otomatis melepas dana ke penjual (*auto-confirmation timeout*) | M |
| BR-02.5 | Pembeli atau penjual dapat membuka sengketa jika ada permasalahan — dana tetap dikunci selama proses mediasi | M |
| BR-02.6 | Seluruh riwayat perubahan status transaksi barang tersimpan sebagai audit trail yang tidak dapat diubah | M |
| BR-02.7 | Platform mendukung transaksi barang digital dengan alur yang disesuaikan (tanpa pengiriman fisik) | S |

---

### BR-03: Escrow & Penyelesaian Pembayaran — Jasa (Milestone)

Platform harus mendukung pembayaran bertahap berbasis pencapaian (*milestone*) untuk transaksi jasa:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-03.1 | Klien dapat mendefinisikan struktur pembayaran bertahap (milestone) saat membuat transaksi jasa | M |
| BR-03.2 | Total nilai kontrak dikunci di escrow di awal — penyedia jasa tidak dapat meminta lebih dari yang disepakati | M |
| BR-03.3 | Setiap milestone memiliki deskripsi, nilai, dan tenggat waktu (*deadline*) yang dicatat dalam sistem | M |
| BR-03.4 | Penyedia jasa mengajukan penyelesaian milestone beserta bukti hasil kerja untuk direview klien | M |
| BR-03.5 | Klien menyetujui milestone untuk memicu pelepasan dana sesuai porsi milestone tersebut | M |
| BR-03.6 | Jika klien tidak merespons dalam batas waktu, sistem otomatis melepas dana milestone (*auto-timeout*) | M |
| BR-03.7 | *Catatan:* Auto-timeout untuk Jasa adalah fitur yang direncanakan — implementasi aktif saat ini hanya mencakup Barang | M |

---

### BR-04: Manajemen Event & Tiket

Platform harus menyediakan sistem pengadaan event end-to-end untuk Event Organizer:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-04.1 | EO dapat membuat event dengan informasi lengkap: nama, tanggal, lokasi, kapasitas, dan kategori tiket | M |
| BR-04.2 | Peserta dapat membeli tiket event secara online — dana masuk ke escrow event | M |
| BR-04.3 | EO dapat memantau penjualan tiket dan anggaran event secara real-time melalui dashboard | M |
| BR-04.4 | Setelah event selesai, platform melepas pendapatan tiket ke EO setelah memotong fee platform | M |
| BR-04.5 | Jika event dibatalkan, sistem otomatis melakukan refund ke seluruh pembeli tiket | M |
| BR-04.6 | EO dapat memantau seluruh pengeluaran vendor dari satu dashboard (*live event budget monitoring*) | M |

---

### BR-05: Vendor Marketplace & Alokasi Multi-Tier

Platform harus menyediakan marketplace vendor yang terintegrasi dengan sistem event:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-05.1 | Vendor dapat mendaftarkan diri dengan profil bisnis, kategori layanan, portofolio, dan harga mulai dari | M |
| BR-05.2 | Platform mendukung multi-tier kategori vendor: Sound System, Catering, Venue, Dekorasi, dan kategori lainnya yang dapat dikonfigurasi | M |
| BR-05.3 | EO dapat browse, filter, dan memilih vendor dari marketplace untuk dialokasikan ke event mereka | M |
| BR-05.4 | EO juga dapat mendaftarkan vendor eksternal (di luar marketplace) — pembayarannya tetap melalui escrow platform | M |
| BR-05.5 | Kontrak vendor terhubung langsung dengan escrow event | M |
| BR-05.6 | Vendor mengajukan pembayaran (*payout request*) secara bertahap dengan upload invoice dan dokumen pendukung | M |
| BR-05.7 | EO mereview dan menyetujui setiap payout request vendor sebelum dana dilepaskan | M |
| BR-05.8 | Seluruh alokasi dan pembayaran vendor tercatat dalam audit trail event | M |

---

### BR-06: Wallet RekberPay & Ledger Keuangan

Platform harus menyediakan dompet digital internal untuk seluruh aktivitas keuangan pengguna:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-06.1 | Setiap pengguna memiliki wallet RekberPay dengan saldo real-time yang akurat | M |
| BR-06.2 | Pengguna dapat mengisi saldo melalui berbagai metode pembayaran via Midtrans (transfer bank, kartu, QRIS) | M |
| BR-06.3 | Pengguna dapat menarik saldo ke rekening bank terdaftar | M |
| BR-06.4 | Sistem secara otomatis menghitung dan memotong: fee platform, fee payment gateway, dan menentukan jumlah bersih (*net payout*) yang diterima penjual/vendor/EO | M |
| BR-06.5 | Seluruh mutasi saldo (debit dan kredit) tercatat dalam ledger keuangan yang dapat diaudit | M |
| BR-06.6 | Saldo tidak dapat negatif dalam kondisi apapun — sistem harus menolak transaksi yang melebihi saldo tersedia | M |
| BR-06.7 | Dana yang sedang dikunci dalam escrow ditampilkan terpisah dari saldo tersedia | M |
| BR-06.8 | Sistem menggunakan mekanisme penguncian (*Serializable Isolation + SELECT FOR UPDATE*) untuk mencegah kondisi balapan (*race condition*) pada mutasi saldo yang bersamaan | M |

---

### BR-07: Audit Log Blockchain (Gasless)

Platform harus mencatat setiap transaksi yang selesai sebagai bukti yang tidak dapat dimanipulasi:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-07.1 | Setiap transaksi yang selesai dicatat di jaringan Avalanche sebagai audit log permanen | M |
| BR-07.2 | Pengguna tidak perlu memiliki wallet kripto atau membayar gas fee — platform menanggung biaya melalui mekanisme relayer | M |
| BR-07.3 | Siapapun dapat memverifikasi keabsahan transaksi melalui bukti on-chain yang dapat diakses publik | M |
| BR-07.4 | Pencatatan on-chain bersifat *append-only* — tidak ada entitas yang dapat mengubah atau menghapus rekaman transaksi | M |
| BR-07.5 | Khusus untuk escrow event bernilai tinggi, audit log on-chain menjadi bukti transparansi bagi semua pihak yang terlibat | M |

---

### BR-08: Verifikasi Identitas (KYC) & Proteksi Fraud

Platform harus memastikan setiap pihak yang berperan sebagai penjual atau penyedia layanan telah terverifikasi:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-08.1 | Pengguna yang ingin menjadi Verified Merchant, Vendor, atau EO wajib menjalani proses KYC | M |
| BR-08.2 | Proses KYC melibatkan upload dokumen identitas (KTP) dan foto selfie untuk verifikasi keaslian | M |
| BR-08.3 | Sistem AI melakukan pre-screening otomatis terhadap dokumen KYC sebelum diteruskan ke Admin | M |
| BR-08.4 | Admin melakukan review final dan memberikan persetujuan atau penolakan KYC | M |
| BR-08.5 | Sistem melakukan penilaian risiko (*fraud scoring*) secara otomatis sebelum melepas dana dari escrow | M |
| BR-08.6 | Transaksi dengan skor risiko tinggi diarahkan ke antrian review manual Admin sebelum diproses | M |
| BR-08.7 | Sistem menerapkan mekanisme *idempotency key* untuk mencegah pemrosesan ganda (*double-spending*) pada setiap operasi keuangan | M |

---

### BR-09: Sengketa & Mediasi

Platform harus menyediakan mekanisme penyelesaian sengketa yang adil dan terstruktur:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-09.1 | Salah satu pihak dalam transaksi dapat membuka sengketa selama transaksi masih aktif | M |
| BR-09.2 | Saat sengketa dibuka, dana escrow dibekukan sampai ada keputusan mediasi | M |
| BR-09.3 | Kedua pihak dapat menyampaikan bukti dan argumen melalui platform | M |
| BR-09.4 | Admin berperan sebagai mediator independen dan memiliki wewenang untuk memutuskan resolusi | M |
| BR-09.5 | Keputusan Admin bersifat final dan dieksekusi otomatis oleh sistem (refund penuh/sebagian, atau dana ke penjual) | M |
| BR-09.6 | Platform mendefinisikan SLA resolusi sengketa yang jelas dan dikomunikasikan ke pengguna | M |

---

### BR-10: Program Loyalitas CRM

Platform harus memberikan insentif kepada pengguna aktif melalui program loyalitas berbasis volume:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-10.1 | Platform mengevaluasi tier loyalitas pengguna setiap bulan berdasarkan volume transaksi | M |
| BR-10.2 | Terdapat empat tier: Bronze, Silver, Gold, Platinum — dengan benefit yang meningkat per tier | M |
| BR-10.3 | Sistem mendistribusikan poin loyalitas secara otomatis berdasarkan aktivitas transaksi | M |
| BR-10.4 | Pengguna menerima notifikasi saat tier berubah | S |
| BR-10.5 | Tier lebih tinggi mendapatkan keuntungan bisnis nyata (fee lebih rendah, prioritas mediasi) | S |

---

### BR-11: Taksonomi & Konfigurasi Platform

Platform harus mendukung pengelolaan kategori dan taksonomi yang fleksibel:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-11.1 | Admin dapat mengelola taksonomi multi-tier untuk: Kategori Barang, Kategori Jasa, Kategori Event, dan Kategori Vendor | M |
| BR-11.2 | Struktur taksonomi dapat dikonfigurasi tanpa perubahan kode (melalui panel admin) | S |
| BR-11.3 | Kategori digunakan untuk filtering di marketplace vendor dan pelaporan bisnis | M |

---

### BR-12: Kapabilitas Administratif

Platform harus memberikan Admin visibilitas dan kontrol penuh atas seluruh operasi platform:

| ID | Kebutuhan | Prioritas |
|----|-----------|-----------|
| BR-12.1 | Admin dapat memantau status seluruh transaksi aktif secara real-time | M |
| BR-12.2 | Admin dapat melakukan *override* pada state transaksi yang bermasalah | M |
| BR-12.3 | Admin memiliki visibilitas ke status background worker (auto-release, CRM evaluasi) dan job scheduling | M |
| BR-12.4 | Admin dapat menangguhkan atau menonaktifkan akun pengguna yang melanggar ketentuan | M |
| BR-12.5 | Seluruh aksi Admin tercatat dalam audit log yang tidak dapat dihapus | M |
| BR-12.6 | Admin dapat mengkonfigurasi fee platform melalui panel tanpa perubahan kode | S |

---

## 8. Business Rules

Business rules adalah aturan yang mengatur cara bisnis beroperasi — berbeda dari requirement yang mendeskripsikan apa yang harus ada.

### BR-RULE-01: Escrow & Aliran Dana

| ID | Aturan |
|----|--------|
| RULE-01.1 | Dana yang masuk ke escrow tidak dapat diakses oleh RekberKuy untuk keperluan operasional — dana escrow bersifat *segregated* |
| RULE-01.2 | Dana hanya dapat dilepas dari escrow melalui salah satu dari empat kondisi: (a) konfirmasi penerima, (b) auto-timeout, (c) keputusan mediasi Admin, (d) pembatalan yang disepakati |
| RULE-01.3 | Fee platform dipotong saat dana dilepaskan dari escrow — bukan saat transaksi dibuat |
| RULE-01.4 | Jika terjadi kegagalan teknis saat melepas dana, sistem harus retry dan tidak boleh kehilangan dana (idempotency wajib) |

### BR-RULE-02: KYC & Eligibilitas

| ID | Aturan |
|----|--------|
| RULE-02.1 | Pengguna dengan status User dapat bertransaksi sebagai pembeli tanpa KYC |
| RULE-02.2 | Untuk bertransaksi sebagai penjual (Verified Merchant), menjual jasa, menjadi EO, atau mendaftar sebagai Vendor — KYC wajib disetujui terlebih dahulu |
| RULE-02.3 | Dokumen KYC yang ditolak hanya dapat diajukan ulang maksimal 3 kali dalam 30 hari |
| RULE-02.4 | Status KYC yang sudah disetujui tidak perlu diulang kecuali ada perubahan identitas atau permintaan Admin |

### BR-RULE-03: Fraud & Risiko

| ID | Aturan |
|----|--------|
| RULE-03.1 | Semua transaksi dengan nilai di atas threshold tertentu wajib melalui fraud scoring sebelum escrow dikunci |
| RULE-03.2 | Transaksi dengan skor risiko TINGGI tidak dapat diproses otomatis — wajib review manual Admin |
| RULE-03.3 | Satu request idempotency key hanya dapat diproses satu kali — duplikat ditolak |
| RULE-03.4 | Akun yang teridentifikasi sebagai fraud dibekukan segera — semua transaksi aktif dihold |

### BR-RULE-04: Vendor & Event

| ID | Aturan |
|----|--------|
| RULE-04.1 | Vendor yang tampil di marketplace wajib telah melalui verifikasi Admin |
| RULE-04.2 | Payout request vendor hanya dapat diproses setelah EO memberikan persetujuan |
| RULE-04.3 | Jika event dibatalkan sebelum berlangsung, seluruh tiket di-refund dan kontrak vendor dibatalkan sesuai ketentuan |
| RULE-04.4 | EO dapat mendaftarkan vendor eksternal, namun pembayaran tetap wajib melalui escrow platform |

### BR-RULE-05: CRM & Loyalitas

| ID | Aturan |
|----|--------|
| RULE-05.1 | Evaluasi tier dilakukan pada hari pertama setiap bulan — berdasarkan GMV bulan sebelumnya |
| RULE-05.2 | Downgrade tier dapat terjadi jika aktivitas turun di bawah threshold tier saat ini |
| RULE-05.3 | Poin loyalitas tidak dapat ditransfer antar akun |

### BR-RULE-06: Fee Structure

| Jenis Transaksi | Fee Platform | Keterangan |
|-----------------|-------------|------------|
| Transaksi Barang | 1.5% dari nilai transaksi | Dipotong saat release |
| Transaksi Jasa | 2.0% dari nilai kontrak | Dipotong di release akhir |
| Tiket Event | 2.0% dari total penjualan tiket | Dipotong saat settlement event |
| Kontrak Vendor | 1.5% dari nilai kontrak vendor | Dipotong per payout yang disetujui |
| Withdraw < Rp 1 Juta | Rp 2.500 flat | Per transaksi withdraw |
| Withdraw ≥ Rp 1 Juta | Gratis | — |

---

## 9. Asumsi & Batasan

### 9.1 Asumsi

| ID | Asumsi |
|----|--------|
| AS-01 | Pengguna memiliki akses internet dan smartphone atau komputer |
| AS-02 | Pengguna memiliki rekening bank Indonesia yang aktif untuk keperluan top-up dan withdraw |
| AS-03 | Midtrans tersedia dan reliable sebagai payment gateway utama |
| AS-04 | Jaringan Avalanche (Fuji Testnet → Mainnet) tersedia untuk pencatatan audit log |
| AS-05 | Groq API tersedia dan memberikan respons dalam batas waktu yang ditentukan untuk KYC dan fraud scoring |
| AS-06 | Pengguna tidak perlu memiliki pengetahuan tentang blockchain atau kripto |
| AS-07 | Seluruh transaksi dalam mata uang Rupiah (IDR) — tidak ada multi-currency di v1.0 |
| AS-08 | Foto selfie + KTP adalah metode KYC yang cukup untuk fase awal |

### 9.2 Batasan

| ID | Batasan |
|----|---------|
| CONST-01 | Platform beroperasi dalam yurisdiksi hukum Indonesia — tunduk pada regulasi BI, OJK, dan UU PDP |
| CONST-02 | Tidak ada pembayaran langsung dengan kripto — platform tidak menyimpan atau mentransfer aset kripto milik pengguna |
| CONST-03 | Blockchain hanya digunakan sebagai audit log — bukan untuk eksekusi logika bisnis atau penyimpanan dana |
| CONST-04 | Data pengguna harus disimpan di server yang berlokasi di wilayah Indonesia (*data residency*) |
| CONST-05 | Tidak ada fitur pinjaman atau kredit di v1.0 — platform murni escrow dan settlement |
| CONST-06 | Mobile app (native iOS/Android) tidak termasuk dalam scope v1.0 |

---

## 10. Dependensi Bisnis

| Dependensi | Tipe | Dampak Jika Tidak Tersedia | Mitigasi |
|------------|------|---------------------------|----------|
| **Midtrans** | Payment Gateway | Pengguna tidak bisa top-up wallet | Siapkan Xendit sebagai fallback |
| **Groq AI** | AI Inference | KYC & fraud scoring tidak berjalan otomatis | Fallback ke review manual Admin sepenuhnya |
| **Avalanche Network** | Blockchain | Audit log on-chain tidak tercatat | Queue transaksi, retry saat network pulih |
| **Supabase** | Database & Storage | Data tidak bisa diakses atau disimpan | Self-hosted PostgreSQL sebagai fallback |
| **Konsultasi Hukum** | Legal | Ketidakpastian compliance OJK/BI | Konsultasi sebelum open beta |
| **Rekening Bisnis** | Finance | Tidak bisa menerima dana dari pengguna | Buka rekening bisnis sebelum go-live |

---

## 11. Risiko Bisnis

| ID | Risiko | Probabilitas | Dampak | Strategi Mitigasi |
|----|--------|-------------|--------|-------------------|
| RISK-01 | Platform dianggap wajib memiliki izin PJSP dari Bank Indonesia | Medium | Kritis | Konsultasi hukum sebelum go-live; pertimbangkan kerjasama dengan PJSP berlisensi |
| RISK-02 | Fraud terorganisir yang memanfaatkan celah sistem | Medium | Tinggi | 3-layer fraud prevention, manual review threshold, KYC ketat |
| RISK-03 | Kepercayaan pengguna rendah di awal | Tinggi | Tinggi | Program referral, KOL marketing, transparansi audit blockchain |
| RISK-04 | Vendor tidak memenuhi kontrak setelah menerima payout | Medium | Tinggi | Escrow payout bertahap, invoice wajib, review EO sebelum release |
| RISK-05 | Sengketa massal pada satu event besar | Rendah | Tinggi | SLA mediasi jelas, team Admin siap standby saat event berskala besar |
| RISK-06 | Keterlambatan go-live karena bottleneck solo developer | Tinggi | Medium | Prioritas ketat sesuai MVP, scope cut agresif jika perlu |
| RISK-07 | Groq API berubah model atau pricing secara signifikan | Medium | Medium | Desain fraud client sebagai interface yang bisa diganti provider |
| RISK-08 | Pengguna menyalahgunakan dispute sebagai alat fraud | Medium | Medium | Batas jumlah dispute aktif, reputasi score berbasis riwayat dispute |

---

## 12. Kriteria Keberhasilan

### 12.1 Business Outcomes per Phase

**Phase 1 — MVP Barang (Q4 2026):**
- [ ] 500 transaksi barang berhasil diselesaikan end-to-end
- [ ] Zero kehilangan dana pengguna akibat bug sistem
- [ ] NPS pengguna beta ≥ 40
- [ ] Waktu resolusi sengketa rata-rata < 48 jam
- [ ] Revenue dari fee transaksi > Rp 5 Juta

**Phase 2 — Jasa & Event (Q1 2027):**
- [ ] 50 EO aktif menggunakan platform
- [ ] 200 vendor terdaftar di marketplace
- [ ] 100% transaksi selesai tercatat di Avalanche Testnet
- [ ] Fraud rate < 0.5% dari total transaksi

**Phase 3 — Full Platform (Q2 2027):**
- [ ] GMV ≥ Rp 1 Miliar/bulan
- [ ] 1.000 MAU aktif
- [ ] Vendor marketplace menjadi sumber 30% dari total GMV
- [ ] Semua audit log di Avalanche Mainnet

### 12.2 Business Rules yang Tidak Boleh Dilanggar

Kondisi berikut dianggap sebagai kegagalan bisnis yang memerlukan penanganan segera:

- ❌ Dana pengguna hilang atau tertukar akibat bug sistem
- ❌ Escrow dilepas tanpa konfirmasi yang sah
- ❌ Data KYC pengguna bocor ke pihak tidak berwenang
- ❌ Platform digunakan untuk aktivitas pencucian uang
- ❌ Sengketa tidak ditangani lebih dari 7 hari kerja

---

## 13. Glossary

| Istilah | Definisi Bisnis |
|---------|-----------------|
| **Escrow** | Mekanisme penitipan dana oleh pihak ketiga (RekberKuy) yang hanya dilepaskan ketika kondisi yang disepakati terpenuhi |
| **Verified Merchant** | Pengguna yang telah melewati proses KYC dan berhak menjual barang atau menawarkan jasa di platform |
| **Event Organizer (EO)** | Pengguna terverifikasi yang membuat dan mengelola event di platform, termasuk pengelolaan tiket dan vendor |
| **Vendor** | Penyedia layanan pendukung event (Sound System, Catering, Venue, Dekorasi, dll) yang terdaftar dan terverifikasi di marketplace |
| **Milestone** | Tahapan pembayaran dalam kontrak jasa — setiap milestone memiliki nilai, deskripsi, dan tenggat waktu |
| **Payout Request** | Pengajuan pencairan dana oleh vendor untuk milestone yang telah diselesaikan, disertai invoice |
| **Auto-Confirmation Timeout** | Mekanisme otomatis yang melepas dana escrow jika penerima tidak memberikan konfirmasi dalam batas waktu |
| **Gasless Transaction** | Pencatatan on-chain yang biaya gas fee-nya ditanggung oleh backend RekberKuy — pengguna tidak perlu membayar |
| **Relayer** | Komponen backend yang bertindak sebagai perantara untuk mengirim transaksi ke blockchain atas nama pengguna |
| **Fraud Scoring** | Penilaian risiko otomatis yang dilakukan AI sebelum dana escrow dilepaskan |
| **Idempotency Key** | Kunci unik per operasi keuangan yang mencegah pemrosesan ganda meskipun request dikirim berulang kali |
| **CRM Loyalty Tiering** | Program loyalitas berbasis volume transaksi dengan 4 tier: Bronze, Silver, Gold, Platinum |
| **Net Payout** | Jumlah bersih yang diterima penjual/vendor/EO setelah dipotong fee platform dan fee payment gateway |
| **Segregated Account** | Rekening terpisah khusus untuk menyimpan dana escrow pengguna — tidak boleh digunakan untuk operasional platform |
| **GMV** | Gross Merchandise Value — total nilai transaksi yang diproses melalui platform |
| **MAU** | Monthly Active Users — pengguna yang melakukan minimal satu transaksi dalam 30 hari terakhir |
| **KYC** | Know Your Customer — proses verifikasi identitas pengguna melalui dokumen resmi |
| **Multi-Tier Vendor Taxonomy** | Sistem kategori vendor berlapis yang mendukung berbagai jenis layanan event (Sound System, Catering, Venue, Dekorasi, dll) |

---

*Dokumen ini adalah **living document** — diperbarui setiap ada keputusan bisnis signifikan, perubahan regulasi, atau pivot strategi. Setiap perubahan material harus tercatat di Version History dengan alasan yang jelas.*

*Untuk spesifikasi teknis, lihat dokumen terkait:*
- *[docs/general-user-guide.md](docs/general-user-guide.md) — panduan pengguna*
- *[docs/application-flow-and-module-guide.md](docs/application-flow-and-module-guide.md) — alur aplikasi detail*
- *[docs/deployment-vps-docker-compose.md](docs/deployment-vps-docker-compose.md) — panduan deployment*
- *[ROADMAP.md](ROADMAP.md) — rencana pengembangan*