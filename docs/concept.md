# 💡 RekberKuy — Product Concept

Dokumen ini menjabarkan konsep bisnis, visi produk, model monetisasi, dan differensiasi RekberKuy sebagai platform escrow modern di Indonesia.

---

## 🌟 Visi & Misi

**Visi:**
> Menjadi infrastruktur kepercayaan digital untuk setiap transaksi di Indonesia maupun luar negeri — dari jual beli barang sehari-hari hingga pengadaan event skala besar.

**Misi:**
- Menghilangkan risiko penipuan dalam transaksi online antara pihak yang belum saling percaya
- Memberikan akses ke sistem escrow yang selama ini hanya ada di transaksi properti & korporat, ke seluruh lapisan masyarakat
- Membangun ekosistem vendor event yang terstruktur, terverifikasi, dan mudah diakses oleh EO

---

## 🎯 Masalah yang Dipecahkan

### Masalah 1: Penipuan Transaksi Online
Transaksi online di Indonesia masih rentan penipuan — penjual menerima pembayaran tapi tidak mengirim barang, atau pembeli menerima barang tapi tidak membayar. Platform marketplace besar sudah mengatasi ini, tapi **transaksi peer-to-peer di luar marketplace** (via WhatsApp, Instagram, Telegram) masih sangat berisiko.

**Solusi RekberKuy:** Dana pembeli ditahan di escrow RekberKuy sampai barang/jasa diterima. Kedua pihak terlindungi.

---

### Masalah 2: Tidak Ada Jaminan di Transaksi Jasa
Freelancer takut tidak dibayar. Klien takut pekerjaan tidak selesai. Tidak ada mekanisme formal yang mudah diakses untuk transaksi jasa peer-to-peer.

**Solusi RekberKuy:** Milestone-based escrow untuk jasa. Dana dicairkan bertahap sesuai kemajuan pekerjaan yang disetujui kedua pihak.

---

### Masalah 3: Pengadaan Event yang Berantakan
EO sering kesulitan mengelola pembayaran ke banyak vendor sekaligus. Vendor khawatir tidak dibayar setelah event. Tidak ada sistem terpusat yang menghubungkan EO dengan vendor terpercaya.

**Solusi RekberKuy:** Platform event dengan vendor marketplace terintegrasi. Semua pembayaran vendor terkelola dalam satu dashboard dengan proteksi escrow.

---

### Masalah 4: Tidak Ada Transparansi Rekam Jejak Transaksi
Jika terjadi sengketa, sulit membuktikan alur transaksi karena semua hanya berdasarkan screenshot chat yang mudah dipalsukan.

**Solusi RekberKuy:** Setiap transaksi dicatat on-chain di Avalanche sebagai audit log yang immutable dan dapat diverifikasi siapapun secara independen.

---

## 🏆 Value Proposition

### Untuk Pembeli
- ✅ Dana aman di escrow — tidak akan hilang jika penjual curang
- ✅ Dispute resolution yang adil dengan mediator
- ✅ Rekam jejak transaksi yang transparan & bisa dibuktikan

### Untuk Penjual / Penyedia Jasa
- ✅ Kepastian pembayaran — pembeli sudah deposit sebelum barang/jasa dikirim
- ✅ Milestone payment untuk pekerjaan bertahap
- ✅ Profil terverifikasi yang meningkatkan kepercayaan

### Untuk Event Organizer (EO)
- ✅ Satu platform untuk kelola tiket + vendor + keuangan event
- ✅ Marketplace vendor terverifikasi — tidak perlu cari vendor dari nol
- ✅ Semua pembayaran vendor terlindungi escrow

### Untuk Vendor
- ✅ Akses ke ratusan EO potensial di satu platform
- ✅ Jaminan pembayaran via escrow — tidak perlu khawatir tidak dibayar
- ✅ Profil & portofolio yang mudah ditemukan EO

---

## 👥 Target Pengguna

### Segmen Primer
| Segmen | Profil | Pain Point Utama |
|--------|--------|-----------------|
| **Reseller / Penjual Online** | Penjual produk di luar marketplace besar, aktif di WA/IG/TikTok | Takut ditipu, tidak punya platform terpercaya |
| **Freelancer & Penyedia Jasa** | Desainer, programmer, fotografer, konsultan | Takut tidak dibayar, tidak ada kontrak formal |
| **Event Organizer** | EO skala menengah (50–1000 pax) | Koordinasi vendor kacau, pembayaran tidak terstruktur |
| **Vendor Event** | Katering, dekorasi, foto/video, sound system | Susah dapat klien baru, takut tidak dibayar |

### Segmen Sekunder
- Komunitas jual beli (grup Facebook, forum, Discord)
- UMKM yang belum masuk marketplace besar
- Panitia acara kampus & komunitas

---


## 💰 Model Monetisasi & Aturan Finansial

RekberKuy menggunakan sistem pembagian keuntungan yang adil dan adaptif, yang dikunci langsung di level core backend engine (`finance_calculator.go`):

### 1. Fee Transaksi Adaptif Melandai (Revenue Utama)
Platform memotong persentase dari setiap dana transaksi yang dicairkan (*RELEASED*) berdasarkan Kasta CRM Loyalty Penjual untuk memicu pertumbuhan ekosistem. Aturan ini berlaku seragam untuk Lini Barang (Goods) dan Jasa (Services):

| Kasta CRM Penjual | Potongan Komisi Platform (Goods & Services) |
|-------------------|---------------------------------------------|
| **NEWBIE** | 10% dari nominal pencairan                  |
| **SILVER** | 6% dari nominal pencairan                   |
| **GOLD** | 3% dari nominal pencairan (Promo)           |



## Sistem Audit Finansial Lini Event (Events Escrow)
Khusus untuk transaksi Event Organizer (EO) skala besar, platform menggunakan mekanika audit pasca-event (*Post-Event Audit Engine*) yang dikunci pada fungsi `CalculateEventAudit`:

- **Komisi Platform Lini Event:** Platform mengunci komisi sebesar **5% di muka (*Upfront Fee*)** dari total dana anggaran event yang terkunci di escrow. Batas maksimal dana yang dapat dibayarkan ke vendor otomatis terpangkas menjadi 95% dari total dana awal. Sisa dana efisiensi dari plafon tersebut akan dihitung sebagai surplus bersih pasca-event.
- **Insentif Surplus EO:** Jika terdapat sisa anggaran (surplus) setelah pelunasan vendor dan pemotongan fee platform, EO berhak mendapatkan bonus insentif dari sisa dana tersebut berdasarkan kasta CRM mereka:

| Skala Dana Event | EO Kasta NEWBIE | EO Kasta SILVER | EO Kasta GOLD |
|------------------|-----------------|-----------------|---------------|
| **Micro Event** ($<$ Rp10 Juta) | 5% dari surplus | 10% dari surplus | 15% dari surplus |
| **Mega Event** ($>$ Rp10 Juta)  | 2% dari surplus | 4% dari surplus  | 8% dari surplus  |\

- **Kebijakan Distribusi Surplus Efisiensi:**
  - **Surplus Mikro ($\le$ Rp 500.000):** Sistem memberikan 100% sisa dana efisiensi langsung ke dompet EO sebagai bonus performa, guna memangkas kompleksitas pembagian dana kecil.
  - **Surplus Makro ($>$ Rp 500.000):** Sisa dana akan dibagi secara adil antara Bonus EO (berdasarkan persentase Kasta CRM) dan Auto-Refund ke saldo RekberPay para peserta event.

- **Proteksi Pengguna (Refund):** Sisa surplus akhir setelah pemotongan jatah bonus EO akan dikembalikan secara otomatis oleh sistem (*Auto-Refund Cluster*) ke saldo RekberPay masing-masing peserta event.

### 2. Biaya Proteksi Awal (Buyer Service Fee)
Pembeli dikenakan biaya proteksi di awal transaksi dengan skema:
- **Kasta Penjual GOLD:** 8% dari nominal dasar.
- **Kasta Penjual SILVER:** 4% dari nominal dasar.
- **Kasta Penjual NEWBIE:** - Menggunakan RekberPay: Rp 2.500 (Promo)
  - Menggunakan Non-RekberPay: Rp 5.000

### 3. Fee Withdraw & Admin Bank
- Biaya penarikan saldo (*Withdraw*) dari dompet RekberPay ke rekening bank pengguna dikenakan biaya flat sebesar **Rp 7.500** (termasuk biaya kliring interbank via payment gateway).

### 4. Fee Withdraw & Admin
- Biaya penarikan saldo (Withdraw) ke rekening bank: **Rp 7.500** flat (termasuk biaya kliring interbank).
- Buyer Service Fee: Dipotong otomatis di awal sebagai biaya proteksi adaptif.

### 5. Fitur Premium (Future)
| Fitur | Harga |
|-------|-------|
| Badge "Verified Seller" prioritas | Rp 99.000 / bulan |
| Vendor featured listing di marketplace | Rp 199.000 / bulan |
| Analytics & laporan transaksi lanjutan | Rp 49.000 / bulan |
| SLA mediasi dispute dipercepat (< 4 jam) | Rp 25.000 / kasus |

---

### 6. B2B / White-Label (Future Phase 5)
Menyediakan escrow-as-a-service untuk platform lain yang ingin mengintegrasikan sistem rekber ke produk mereka. Revenue dari lisensi API.

---

## 🔄 Alur Dana (Money Flow)

```
Pembeli Top-Up
      │
      ▼
 Wallet Pembeli
      │
      │ (saat buat transaksi)
      ▼
  Escrow RekberKuy ─────────────────────► Platform Fee (dipotong saat selesai)
      │
      │ (setelah konfirmasi selesai)
      ▼
 Wallet Penjual / Penyedia Jasa / EO / Vendor
      │
      │ (saat withdraw)
      ▼
 Rekening Bank Pengguna
```

---

## 🆚 Kompetitor & Differensiasi

| Fitur | RekberKuy | Rekber Manual (WA) | Marketplace (Tokopedia, dll) | Platform Freelance |
|-------|-----------|-------------------|------------------------------|-------------------|
| Rekber Barang | ✅ | ✅ (manual, rawan) | ✅ | ❌ |
| Rekber Jasa | ✅ | ⚠️ (tidak terstruktur) | ❌ | ✅ (terbatas) |
| Rekber Event + Tiket | ✅ | ❌ | ❌ | ❌ |
| Vendor Marketplace | ✅ | ❌ | ❌ | ❌ |
| Audit Log Blockchain | ✅ | ❌ | ❌ | ❌ |
| Milestone Payment | ✅ | ❌ | ❌ | ✅ |
| Dispute Resolution | ✅ (terstruktur) | ❌ | ✅ | ✅ |
| KYC Terverifikasi | ✅ | ❌ | ✅ | ✅ |
| Open API / Integrasi | ✅ (roadmap) | ❌ | ❌ | ❌ |

**Keunggulan utama RekberKuy:**
1. **Satu-satunya** platform yang mengcover barang, jasa, DAN event dalam satu ekosistem
2. **Vendor marketplace** khusus event — belum ada yang fokus di segmen ini
3. **Blockchain audit log** — transparansi yang tidak bisa dipalsukan

---

## 🔐 Keamanan & Kepercayaan

### Lapisan Keamanan
1. **KYC** — semua penjual, EO, dan vendor wajib verifikasi identitas
2. **Escrow** — dana tidak pernah langsung berpindah dari pembeli ke penjual
3. **Dispute Resolution** — mediator manusia (Admin) dengan SLA jelas
4. **Blockchain Audit** — rekam jejak transaksi immutable di Avalanche
5. **Enkripsi** — semua data sensitif dienkripsi at-rest dan in-transit
6. **AI Multimodal Guardian (Anti-Review Bombing)**
RekberKuy mengintegrasikan AI Multimodal (Gemini API) sebagai lapis verifikasi ulasan negatif:
- Pembeli yang memberikan rating < 3 Bintang **wajib** menyertakan bukti Video Unboxing tanpa terpotong (*uncut*).
- AI akan menganalisis integritas video, mendeteksi *editing cuts*, dan mencocokkan nomor resi kurir sebelum rating diizinkan memotong reputasi merchant di sistem CRM.

### Proteksi Pengguna
- Dana di escrow **bukan** milik RekberKuy — tidak bisa digunakan untuk operasional
- Jika terjadi dispute, dana **difreeze** sampai ada keputusan mediator
- Riwayat transaksi tersimpan minimal 5 tahun (compliance)

---

## 📐 Konsep UX Utama

### Prinsip Desain
1. **Trust First** — setiap elemen UI harus membangun kepercayaan (badge, rating, verifikasi)
2. **Simplicity** — alur transaksi maksimal 5 langkah dari mulai hingga selesai
3. **Transparency** — pengguna selalu tahu di mana dana mereka berada
4. **Mobile-First** — mayoritas pengguna akses via smartphone

### Status Transaksi (Universal)
Semua lini bisnis (Barang, Jasa, Event) dipaksa mematuhi aliran perubahan status (*State Machine*) yang kaku di level database:

WAITING_PAYMENT ──(User Bayar)──► FUNDS_LOCKED ──(Konfirmasi/Selesai)──► RELEASED (Sukses)
                                       │
                         (Jika Ada Sengketa/Komplain)
                                       ▼
                                   DISPUTED ──(Resolusi AI/Admin)──► REFUNDED (Dana Kembali)

---

### 7. Multi-Currency Escrow (Future Phase 5)
Membuka peluang integrasi dompet aset digital (Stablecoin seperti USDT/USDC) khusus untuk transaksi lintas negara (*cross-border freelancing*) dengan tetap mematuhi koridor regulasi lokal yang berlaku.

## 🚀 Go-To-Market Strategy

### Phase 1: Community-Led Growth
- Target komunitas jual beli di Facebook, Telegram, Discord
- Endorsement dari influencer teknologi & bisnis online
- Fitur referral — pengguna dapat fee cashback jika ajak teman

### Phase 2: EO & Vendor Network
- Partnership dengan komunitas EO (IECA, forum EO lokal)
- Onboarding massal vendor event di 5 kota besar
- Case study dari EO yang berhasil pakai RekberKuy

### Phase 3: B2B & API
- Tawarkan API escrow ke platform jual beli komunitas
- White-label untuk bank / fintech yang ingin layanan escrow

---

*Dokumen ini adalah living document — update saat ada perubahan strategi atau pivot produk.*