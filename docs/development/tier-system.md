# Tier & Subsidi Silang System
# RekberKuy Platform

> Di RekberKuy, **kepercayaan dan reputasi adalah aset**.
> Sistem ini memberikan reward kepada pengguna berkualitas dan
> membebankan lebih kepada pengguna berisiko — secara otomatis.

---

## Filosofi: Subsidi Silang

```
Seller bereputasi buruk → bayar fee LEBIH TINGGI
        |
        | fee tinggi ini membiayai...
        ↓
Buyer loyal → dapat DISKON + CASHBACK
Seller bereputasi baik → bayar fee LEBIH RENDAH
```

**Dampak yang diharapkan:**
- Seller produk abal-abal tidak bertahan lama — biaya platform makin mahal
- Seller berkualitas makin kompetitif — biaya makin murah seiring reputasi naik
- Buyer loyal makin betah — makin sering transaksi, makin banyak benefit
- Platform tetap sustain — revenue dari seller bereputasi rendah mensubsidi diskon

---

## Bagian 1: Seller Reputation Tier

Berlaku untuk: **Verified Merchant, Event Organizer, Vendor**

### Cara Hitung Rating

- Rating dihitung **per bulan** dari rata-rata semua bintang yang diterima bulan itu
- Skala: 1–5 bintang, threshold positif = **≥ 3 bintang**
- Pembeli/klien/EO dapat memberikan rating setelah transaksi selesai
- Evaluasi dilakukan **hari pertama setiap bulan** via `crm_worker`

### Grace Period (Seller Baru)

```
Hari 1 pendaftaran → Grace Period 30 hari → Fee Standard (1.5%)
                                │
                    ┌───────────┴────────────┐
                    │                        │
              Dapat rating              Tidak ada rating
              dalam 30 hari            dalam 30 hari
                    │                        │
                    ▼                        ▼
            Dievaluasi normal          Tetap Standard (1.5%)
                                       sampai rating pertama masuk
```

### Tabel Reputation Tier

| Tier | Rating Bulanan | Fee Platform | Status |
|------|---------------|-------------|--------|
| 🏆 **Elite** | ≥ 4.5 ⭐ | **0.75%** | Terbaik — disubsidi sistem |
| ✅ **Trusted** | 3.5 – 4.4 ⭐ | **1.0%** | Di atas rata-rata |
| 🆕 **Standard** | 3.0 – 3.4 ⭐ atau grace period | **1.5%** | Baseline |
| ⚠️ **Developing** | 2.0 – 2.9 ⭐ | **2.0%** | Di bawah ekspektasi |
| 🚨 **At Risk** | < 2.0 ⭐ | **2.5%** | Perlu perbaikan segera |
| ❌ **Suspended** | Pelanggaran berulang | N/A | Tidak bisa bertransaksi |

### Aturan Naik/Turun Tier

```
NAIK:
  - Rating bulan ini ≥ threshold tier di atas → naik 1 tier
  - Tidak bisa loncat 2 tier sekaligus (Elite hanya dari Trusted)

TURUN:
  - Rating bulan ini < threshold tier saat ini → turun 1 tier
  - Kecuali At Risk → bisa langsung Suspended jika ada dispute aktif

FREEZE:
  - Tier tidak berubah jika tidak ada transaksi bulan itu
    (rating bulan lalu dicarry over)
  - Seller tidak dihukum karena tidak aktif sementara

SUSPENDED:
  - 3x berturut-turut At Risk
  - Atau: dispute kalah 2x dalam 1 bulan
  - Atau: terbukti fraud
  - Butuh review Admin untuk aktif kembali
```

---

## Bagian 2: Buyer CRM Loyalty Tier

Berlaku untuk: **User (Pembeli)**

### Cara Hitung Tier

- Berdasarkan **total GMV sebagai pembeli** dalam 1 bulan
- Evaluasi otomatis **hari pertama setiap bulan** via `crm_worker`
- Tier berlaku untuk seluruh bulan berikutnya

### Tabel CRM Loyalty Tier

| Tier | GMV/bulan | Fee Discount | Cashback | Benefit Tambahan |
|------|-----------|-------------|---------|-----------------|
| 🥉 **Bronze** | < Rp 1 Juta | 0% | 0% | Akses standar |
| 🥈 **Silver** | Rp 1–10 Juta | Diskon Rp 2.500/tx | 0.5% ke RekberPay | Support lebih cepat |
| 🥇 **Gold** | Rp 10–50 Juta | Diskon Rp 5.000/tx | 1.0% ke RekberPay | Prioritas mediasi |
| 💎 **Platinum** | > Rp 50 Juta | Gratis biaya withdraw | 1.5% ke RekberPay | Dedicated support |

### Penjelasan Benefit

**Fee Discount:**
- Potongan dari biaya layanan per transaksi yang ditanggung pembeli
- Dihitung saat checkout, langsung mengurangi jumlah yang dibayar
- Bukan diskon dari harga barang/jasa — ini subsidi dari platform

**Cashback ke RekberPay:**
- % dari nilai transaksi bruto dikreditkan ke saldo RekberPay
- Dikreditkan saat transaksi berstatus **COMPLETED** (bukan saat dibuat)
- Cashback dibatalkan jika transaksi di-dispute dan buyer kalah
- Cashback tidak berlaku untuk transaksi yang dibatalkan

**Contoh Buyer Gold:**
```
Beli barang Rp 500.000
Fee layanan normal: Rp 0 (fee ditanggung seller di RekberKuy)
Cashback 1.0%: +Rp 5.000 ke RekberPay wallet
Fee withdraw jika ada: diskon Rp 5.000

Efek nyata: setiap Rp 500.000 belanja, dapat Rp 5.000 kembali
```

---

## Bagian 3: Mekanisme Subsidi Silang

### Simulasi Revenue per Transaksi Rp 1.000.000

```
SKENARIO A — Seller At Risk, Buyer Bronze (paling buruk)
Seller fee: 2.5% = Rp 25.000 → masuk platform
Buyer cashback: 0
Buyer fee discount: 0
Platform NET: Rp 25.000

SKENARIO B — Seller Standard, Buyer Silver (baseline)
Seller fee: 1.5% = Rp 15.000 → masuk platform
Buyer cashback: 0.5% = Rp 5.000 → keluar dari platform
Buyer fee discount: Rp 2.500 → keluar dari platform
Platform NET: Rp 7.500

SKENARIO C — Seller Elite, Buyer Platinum (paling baik)
Seller fee: 0.75% = Rp 7.500 → masuk platform
Buyer cashback: 1.5% = Rp 15.000 → keluar dari platform
Buyer fee discount: withdraw gratis (estimasi Rp 2.500) → keluar
Platform NET: -Rp 10.000 ← NEGATIF

⚠️  Skenario C tidak boleh terlalu sering! Ini yang perlu diperhatikan.
```

### Sustainability Check

Skenario C (Elite + Platinum) menghasilkan revenue negatif per transaksi. Ini **acceptable** karena:
1. Seller Elite dan Buyer Platinum adalah pengguna PALING AKTIF — volume transaksi mereka sangat tinggi
2. Revenue dari Skenario A (At Risk + Bronze) mensubsidi kerugian dari Skenario C
3. Cashback dikreditkan ke RekberPay — uangnya **tetap di ekosistem platform** (hanya bisa dipakai di RekberKuy)
4. Platform mendapat float benefit dari dana di wallet yang belum dicairkan

### Batas Proteksi

Agar platform tidak rugi berlebihan, tambahkan aturan:
```
RULE-TIER-01: Cashback per bulan maksimal Rp 500.000 per user
              (mencegah abuse — user yang buka banyak akun)
RULE-TIER-02: Cashback hanya untuk transaksi dengan seller Trusted ke atas
              (cashback tidak berlaku jika beli dari seller At Risk)
RULE-TIER-03: Cashback expire dalam 6 bulan jika tidak digunakan
              (mendorong user tetap aktif)
RULE-TIER-04: Seller Elite harus punya minimal 10 transaksi bulan itu
              (mencegah seller dengan 1 transaksi sempurna dapat fee 0.75%)
```

---

## Bagian 4: Fee Matrix Lengkap

Fee yang seller bayar berdasarkan kombinasi domain transaksi + tier:

| Domain | Elite (0.75%) | Trusted (1.0%) | Standard (1.5%) | Developing (2.0%) | At Risk (2.5%) |
|--------|-------------|----------------|-----------------|------------------|----------------|
| **Barang** | Rp 3.750 | Rp 5.000 | Rp 7.500 | Rp 10.000 | Rp 12.500 |
| **Jasa** | Rp 7.500 | Rp 10.000 | Rp 15.000 | Rp 20.000 | Rp 25.000 |
| **Event (tiket)** | Rp 7.500 | Rp 10.000 | Rp 15.000 | Rp 20.000 | Rp 25.000 |
| **Vendor kontrak** | Rp 3.750 | Rp 5.000 | Rp 7.500 | Rp 10.000 | Rp 12.500 |

*(contoh untuk transaksi Rp 500.000)*

---

## Bagian 5: Domain Model Update

### Update `domain/profile.go`

```go
type SellerTier string
const (
    SellerTierGracePeriod SellerTier = "GRACE_PERIOD"
    SellerTierElite       SellerTier = "ELITE"       // fee 0.75%
    SellerTierTrusted     SellerTier = "TRUSTED"     // fee 1.0%
    SellerTierStandard    SellerTier = "STANDARD"    // fee 1.5%
    SellerTierDeveloping  SellerTier = "DEVELOPING"  // fee 2.0%
    SellerTierAtRisk      SellerTier = "AT_RISK"     // fee 2.5%
    SellerTierSuspended   SellerTier = "SUSPENDED"
)

type BuyerTier string
const (
    BuyerTierBronze   BuyerTier = "BRONZE"
    BuyerTierSilver   BuyerTier = "SILVER"
    BuyerTierGold     BuyerTier = "GOLD"
    BuyerTierPlatinum BuyerTier = "PLATINUM"
)

type Profile struct {
    UserID       string
    Name         string
    Phone        string
    // Buyer CRM
    BuyerTier       BuyerTier
    BuyerGMVMonth   int64     // GMV bulan ini sebagai pembeli
    CashbackBalance int64     // saldo cashback pending (dikreditkan saat tx COMPLETED)
    // Seller Reputation
    SellerTier         SellerTier
    SellerRatingMonth  float64   // rata-rata rating bulan ini
    SellerRatingCount  int       // jumlah rating bulan ini
    SellerTxCount      int       // jumlah transaksi bulan ini
    GracePeriodEndsAt  *time.Time // nil jika bukan grace period lagi
}
```

### Update `domain/finance.go`

```go
type FeeConfig struct {
    // Seller fee per tier
    SellerFeeRate map[SellerTier]float64
    // Buyer cashback per tier
    BuyerCashbackRate map[BuyerTier]float64
    // Buyer fee discount per transaksi (flat, IDR)
    BuyerFeeDiscount map[BuyerTier]int64
    // Batas cashback per bulan
    MaxCashbackPerMonth int64
    // Minimum transaksi untuk dapat tier Elite
    EliteMinTxCount int
}

// FeeCalculation hasil kalkulasi untuk 1 transaksi
type FeeCalculation struct {
    GrossAmount    int64
    SellerFeeRate  float64
    PlatformFee    int64   // dipotong dari seller
    SellerPayout   int64   // grossAmount - platformFee
    BuyerCashback  int64   // dikreditkan ke wallet buyer saat COMPLETED
    BuyerDiscount  int64   // langsung mengurangi tagihan buyer
}
```

### Event Baru untuk EDD

```go
// Tambahkan ke domain events:
SellerTierChanged   { userId, fromTier, toTier, avgRating, txCount, month }
BuyerTierChanged    { userId, fromTier, toTier, gmvMonth, month }
CashbackCredited    { userId, txId, amount, newBalance }
CashbackExpired     { userId, amount, expiredAt }
SellerSuspended     { userId, reason, adminId }
SellerReinstated    { userId, adminId }
```

---

## Bagian 6: Update `crm_worker.go`

Worker yang sudah ada perlu diperluas untuk handle dua evaluasi:

```go
// Pseudocode — implementasi di internal/worker/crm_worker.go

func (w *CRMWorker) RunMonthlyEvaluation(ctx context.Context) error {
    // 1. Evaluasi Buyer Tier
    buyers, _ := w.userRepo.GetAllActiveBuyers(ctx)
    for _, buyer := range buyers {
        gmv := w.txRepo.GetBuyerGMVLastMonth(ctx, buyer.UserID)
        newTier := calculateBuyerTier(gmv)
        if newTier != buyer.BuyerTier {
            w.profileRepo.UpdateBuyerTier(ctx, buyer.UserID, newTier)
            w.eventBus.Publish(BuyerTierChanged{...})
        }
    }

    // 2. Evaluasi Seller/Vendor/EO Tier
    sellers, _ := w.userRepo.GetAllActiveSellers(ctx)
    for _, seller := range sellers {
        // Cek grace period
        if seller.IsInGracePeriod() {
            continue // skip evaluasi, biarkan grace period berjalan
        }

        avgRating := w.ratingRepo.GetMonthlyAvgRating(ctx, seller.UserID)
        txCount := w.txRepo.GetSellerTxCountLastMonth(ctx, seller.UserID)
        newTier := calculateSellerTier(avgRating, txCount)

        if newTier != seller.SellerTier {
            w.profileRepo.UpdateSellerTier(ctx, seller.UserID, newTier)
            w.eventBus.Publish(SellerTierChanged{...})
        }
    }

    // 3. Expire cashback yang sudah 6 bulan
    w.walletRepo.ExpireOldCashback(ctx, time.Now().AddDate(0, -6, 0))

    return nil
}

func calculateSellerTier(avgRating float64, txCount int) SellerTier {
    switch {
    case avgRating >= 4.5 && txCount >= 10: return SellerTierElite
    case avgRating >= 3.5:                  return SellerTierTrusted
    case avgRating >= 3.0:                  return SellerTierStandard
    case avgRating >= 2.0:                  return SellerTierDeveloping
    default:                                return SellerTierAtRisk
    }
}

func calculateBuyerTier(gmvIDR int64) BuyerTier {
    switch {
    case gmvIDR >= 50_000_000: return BuyerTierPlatinum
    case gmvIDR >= 10_000_000: return BuyerTierGold
    case gmvIDR >= 1_000_000:  return BuyerTierSilver
    default:                   return BuyerTierBronze
    }
}
```

---

## Bagian 7: Tabel Baru yang Dibutuhkan di Database

```sql
-- Rating per transaksi
CREATE TABLE transaction_ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id  UUID NOT NULL REFERENCES transactions(id),
    rater_id        UUID NOT NULL,    -- siapa yang kasih rating
    ratee_id        UUID NOT NULL,    -- siapa yang dirating
    score           SMALLINT NOT NULL CHECK (score BETWEEN 1 AND 5),
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (transaction_id, rater_id)  -- 1 orang 1 rating per transaksi
);

-- Snapshot tier bulanan (untuk audit & dispute tier)
CREATE TABLE tier_snapshots (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    tier_type   VARCHAR(10) NOT NULL,  -- 'BUYER' | 'SELLER'
    tier_value  VARCHAR(20) NOT NULL,
    avg_rating  NUMERIC(3,2),          -- null untuk buyer tier
    gmv_month   BIGINT,                -- null untuk seller tier
    tx_count    INT,
    evaluated_at TIMESTAMPTZ NOT NULL,
    month_year  CHAR(7) NOT NULL       -- format: '2026-10'
);

-- Cashback ledger
CREATE TABLE cashback_ledger (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    transaction_id  UUID NOT NULL,
    amount          BIGINT NOT NULL,
    status          VARCHAR(10) NOT NULL,  -- PENDING | CREDITED | EXPIRED | REVERSED
    credited_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Bagian 8: Tampilan ke Pengguna

### Yang Ditampilkan di Dashboard

**Untuk Buyer:**
```
Tier kamu: 🥇 Gold
GMV bulan ini: Rp 12.500.000 / Rp 50.000.000 untuk Platinum
Cashback bulan ini: Rp 45.000
Cashback pending (belum dikreditkan): Rp 12.000
Fee discount per transaksi: Rp 5.000
Cashback kadaluarsa pada: April 2027
```

**Untuk Seller:**
```
Reputasi kamu: ✅ Trusted
Rating rata-rata bulan ini: 4.2 ⭐ (dari 23 transaksi)
Fee platform kamu: 1.0% (hemat 0.5% dari Standard)
Naik ke Elite: butuh rating ≥ 4.5 dan minimal 10 transaksi
```

---

## BRD & DDD Update Checklist

Setelah dokumen ini dibuat, update bagian berikut:

**BRD.md:**
- [ ] BR-10 (CRM Loyalty) → replace dengan sistem dual-tier ini
- [ ] Appendix B (Fee Structure) → tambah tabel seller tier fee + buyer cashback
- [ ] BR-RULE-05 (CRM Rules) → tambah semua RULE-TIER-*

**DDD.md:**
- [ ] Tambah `SellerTier` dan `BuyerTier` sebagai Value Object
- [ ] Update aggregate `User/Profile` dengan field baru
- [ ] Tambah `Rating` sebagai Entity baru
- [ ] Update `Finance` dengan `FeeCalculation` yang tier-aware
- [ ] Tambah domain events baru di atas

**EDD.md:**
- [ ] Tambah event flow untuk: rating diberikan → tier dievaluasi → tier berubah
- [ ] Tambah cashback event flow
- [ ] Update crm_worker event

**BDD.md:**
- [ ] Tambah Feature: Rating System
- [ ] Tambah Feature: Tier Evaluation & Subsidi Silang
- [ ] Tambah Feature: Cashback Flow