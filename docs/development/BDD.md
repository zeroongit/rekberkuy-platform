# Behavior-Driven Development (BDD)
# RekberKuy Platform

> BDD adalah jembatan antara DDD/EDD (domain & event) dan TDD (test code).
> Setiap scenario di sini adalah acceptance criteria yang harus dipenuhi
> sebelum fitur dianggap selesai.

---

## Format

Gunakan Gherkin — bahasa yang bisa dibaca developer DAN stakeholder non-teknis:

```
Feature: [Nama fitur — domain dari DDD]

  Background:
    Given [kondisi awal yang berlaku untuk semua scenario]

  Scenario: [nama scenario — happy path atau edge case]
    Given [kondisi awal]
    When  [aksi yang dilakukan]
    Then  [hasil yang diharapkan]
    And   [hasil tambahan]
```

---

## Feature: Transaksi Barang (Goods Escrow)

```gherkin
Feature: Escrow untuk transaksi barang

  Background:
    Given pembeli "budi@test.com" memiliki saldo RekberPay Rp 1.000.000
    And penjual "sari@test.com" adalah Verified Merchant
    And penjual sudah KYC terverifikasi

  Scenario: Pembeli berhasil membuat transaksi dan escrow terkunci
    Given pembeli ingin membeli "Laptop Bekas" seharga Rp 500.000
    When pembeli membuat transaksi baru
    Then status transaksi adalah "ACTIVE"
    And saldo available pembeli berkurang Rp 500.000
    And saldo locked pembeli bertambah Rp 500.000
    And penjual menerima notifikasi transaksi baru

  Scenario: Pembeli konfirmasi penerimaan barang
    Given transaksi "TRX-001" berstatus "ACTIVE"
    And penjual sudah input nomor resi "JNE123456"
    When pembeli mengkonfirmasi penerimaan barang
    Then status transaksi berubah menjadi "COMPLETED"
    And saldo available penjual bertambah Rp 492.500 (setelah fee 1.5%)
    And fee platform Rp 7.500 tercatat di ledger platform
    And transaksi dicatat sebagai audit log di blockchain

  Scenario: Auto-release setelah 3 hari timeout
    Given transaksi "TRX-002" berstatus "ACTIVE"
    And penjual sudah input nomor resi
    And sudah melewati 3 hari sejak nomor resi diinput
    When auto_release_worker berjalan
    Then status transaksi berubah menjadi "COMPLETED"
    And dana dilepaskan ke penjual secara otomatis
    And pembeli menerima notifikasi auto-release

  Scenario: Pembeli gagal buat transaksi karena saldo tidak cukup
    Given pembeli memiliki saldo RekberPay Rp 100.000
    When pembeli mencoba membuat transaksi senilai Rp 500.000
    Then sistem menolak transaksi dengan error "INSUFFICIENT_BALANCE"
    And saldo pembeli tidak berubah
    And tidak ada escrow yang terkunci

  Scenario: Transaksi diblokir karena fraud score tinggi
    Given transaksi bernilai Rp 50.000.000 dari user baru (3 hari)
    When sistem mengirim request ke fraud scoring service
    And fraud score yang dikembalikan adalah 0.85
    And FRAUD_UNSAFE_THRESHOLD adalah 0.5
    Then sistem menolak mengunci escrow
    And transaksi masuk antrian review manual Admin
    And pembeli menerima notifikasi bahwa transaksi sedang direview
```

---

## Feature: Transaksi Jasa — Milestone Payment

```gherkin
Feature: Escrow bertahap untuk transaksi jasa

  Background:
    Given klien "andi@test.com" memiliki saldo Rp 3.000.000
    And penyedia jasa "maya@test.com" adalah Verified Merchant

  Scenario: Klien membuat kontrak jasa dengan 3 milestone
    Given klien ingin memesan jasa desain logo seharga Rp 3.000.000
    And milestone: [Konsep Rp1jt, Revisi Rp1jt, Final Rp1jt]
    When penyedia jasa menerima kontrak
    And klien mengunci total dana
    Then saldo klien berkurang Rp 3.000.000
    Dan seluruh dana terkunci di escrow
    And status transaksi adalah "ACTIVE"

  Scenario: Penyedia jasa submit milestone dan klien approve
    Given transaksi jasa "TRX-SVC-001" berstatus "ACTIVE"
    And milestone pertama (Rp 1.000.000) belum selesai
    When penyedia jasa menandai milestone selesai dengan bukti hasil kerja
    And klien menyetujui milestone pertama
    Then saldo available penyedia jasa bertambah Rp 1.000.000
    And milestone pertama berstatus "PAID"
    And 2 milestone berikutnya masih PENDING

  Scenario: Auto-release milestone jika klien tidak respons 3 hari
    Given milestone kedua sudah disubmit penyedia jasa
    And klien tidak memberikan respons selama 3 hari
    When auto_release_worker berjalan
    Then milestone kedua otomatis disetujui
    And dana milestone kedua dilepas ke penyedia jasa

  Scenario: Klien minta revisi sebelum approve milestone
    Given milestone pertama sudah disubmit penyedia jasa
    When klien meminta revisi dengan catatan "Logo kurang sesuai brief"
    Then milestone kembali ke status "REVISION"
    And penyedia jasa menerima notifikasi permintaan revisi
    And dana milestone belum dilepas
```

---

## Feature: Wallet — Top-Up & Withdraw

```gherkin
Feature: Manajemen saldo RekberPay

  Scenario: Top-up berhasil via transfer bank Midtrans
    Given pengguna "budi@test.com" memiliki saldo Rp 0
    When pengguna melakukan top-up Rp 500.000 via transfer bank
    And Midtrans mengirim webhook konfirmasi pembayaran
    And signature webhook valid
    Then saldo pengguna bertambah Rp 500.000
    And ledger mencatat CREDIT Rp 500.000
    And pengguna menerima notifikasi top-up berhasil

  Scenario: Webhook Midtrans dengan signature tidak valid ditolak
    Given webhook diterima dari Midtrans
    When signature webhook tidak valid
    Then sistem menolak request dengan HTTP 401
    And saldo tidak berubah
    And kejadian ini dicatat di log security

  Scenario: Top-up duplikat dengan idempotency key yang sama ditolak
    Given top-up "TOPUP-001" sudah berhasil diproses
    When webhook yang sama dikirim ulang dengan order_id "TOPUP-001"
    Then sistem mengembalikan respons sukses (200) tanpa memproses ulang
    And saldo tidak berubah dua kali

  Scenario: Pengguna mengajukan withdraw
    Given pengguna memiliki saldo available Rp 500.000
    And pengguna sudah mendaftarkan rekening BCA 1234567890
    When pengguna mengajukan withdraw Rp 200.000
    Then status withdraw menjadi "PENDING_APPROVAL"
    And saldo available berkurang Rp 200.000
    And saldo locked bertambah Rp 200.000
    And Admin mendapat notifikasi ada permintaan withdraw baru

  Scenario: Admin approve withdraw
    Given ada permintaan withdraw Rp 200.000 berstatus "PENDING_APPROVAL"
    When Admin menyetujui permintaan withdraw
    Then status withdraw menjadi "COMPLETED"
    And saldo locked berkurang Rp 200.000
    And ledger mencatat DEBIT Rp 200.000
    And pengguna menerima notifikasi withdraw diproses
```

---

## Feature: KYC — Verifikasi Identitas

```gherkin
Feature: Proses KYC untuk upgrade peran pengguna

  Scenario: User submit KYC dan AI pre-screening lolos
    Given pengguna "budi@test.com" berstatus USER (belum KYC)
    When pengguna mengupload foto KTP dan selfie yang jelas
    And AI pre-screening menghasilkan score 0.85
    Then status KYC menjadi "PENDING" (menunggu review Admin)
    And Admin mendapat notifikasi KYC baru untuk direview

  Scenario: Admin approve KYC
    Given KYC pengguna berstatus "PENDING"
    When Admin menyetujui KYC
    Then status KYC menjadi "VERIFIED"
    And peran pengguna berubah dari USER ke VERIFIED_MERCHANT
    And pengguna menerima email konfirmasi KYC disetujui

  Scenario: Admin reject KYC karena foto tidak jelas
    Given KYC pengguna berstatus "PENDING"
    When Admin menolak KYC dengan alasan "Foto KTP tidak terbaca"
    Then status KYC menjadi "REJECTED"
    And pengguna menerima email penolakan dengan alasan
    And pengguna masih bisa mengajukan ulang (maksimal 3x dalam 30 hari)

  Scenario: User coba transaksi sebagai penjual tanpa KYC
    Given pengguna belum melewati KYC (status UNVERIFIED)
    When pengguna mencoba membuat listing produk
    Then sistem menolak dengan error "KYC_REQUIRED"
    Dan pengguna diarahkan ke halaman KYC
```

---

## Feature: Dispute & Mediasi

```gherkin
Feature: Penyelesaian sengketa transaksi

  Background:
    Given transaksi "TRX-001" berstatus "ACTIVE"
    And escrow terkunci Rp 500.000

  Scenario: Pembeli membuka dispute
    When pembeli membuka sengketa dengan alasan "Barang tidak sesuai deskripsi"
    And pembeli mengupload foto bukti
    Then status transaksi berubah menjadi "DISPUTED"
    And escrow difreeze (tidak bisa dilepas otomatis)
    And penjual mendapat notifikasi dan 48 jam untuk merespons
    And Admin mendapat notifikasi dispute baru

  Scenario: Admin memutuskan refund penuh ke pembeli
    Given dispute "DISP-001" berstatus aktif
    And Admin sudah mereview bukti dari kedua pihak
    When Admin memutuskan refund penuh ke pembeli
    Then saldo available pembeli bertambah Rp 500.000
    Dan saldo locked kembali ke 0
    And status transaksi berubah menjadi "RESOLVED"
    And kedua pihak menerima notifikasi keputusan

  Scenario: Admin memutuskan split — 60% pembeli, 40% penjual
    Given dispute "DISP-002" berstatus aktif
    When Admin memutuskan split pembayaran
    Then saldo available pembeli bertambah Rp 300.000 (60%)
    And saldo available penjual bertambah Rp 192.500 (40% - fee 1.5%)
    And status transaksi berubah menjadi "RESOLVED"
```

---

## Feature: Vendor Marketplace & Event

```gherkin
Feature: Vendor marketplace untuk Event Organizer

  Background:
    Given EO "andi@test.com" sudah KYC terverifikasi
    And Vendor "katering-maya" sudah terverifikasi

  Scenario: EO memilih vendor dari marketplace untuk event
    Given EO sudah membuat event "Seminar Tech 2026"
    When EO browse marketplace dan memilih "katering-maya"
    And EO membuat kontrak vendor senilai Rp 5.000.000
    Then kontrak vendor dibuat dengan status "PENDING_PAYMENT"
    And EO mengunci Rp 5.000.000 ke escrow vendor

  Scenario: Vendor submit payout request dengan invoice
    Given kontrak vendor sudah aktif
    When vendor mengupload invoice dan menandai pekerjaan selesai
    Then payout request berstatus "PENDING_APPROVAL"
    And EO mendapat notifikasi untuk mereview invoice

  Scenario: EO approve payout request vendor
    Given payout request vendor berstatus "PENDING_APPROVAL"
    When EO menyetujui payout request
    Then saldo available vendor bertambah Rp 5.000.000 (dikurangi fee 1.5%)
    And fee Rp 75.000 masuk ke platform
    And transaksi vendor dicatat di blockchain
```

---

## Cara Implementasi BDD ke Kode

### Untuk Go Backend (menggunakan godog)
```go
// e2e-qa/features/goods_escrow.feature → implementasi di:
// apps/core-service/internal/usecase/transaction_goods_usecase_test.go

func TestGoodsEscrowScenarios(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format: "pretty",
            Paths:  []string{"features/goods"},
        },
    }
    suite.Run()
}
```

### Untuk Frontend (menggunakan Playwright)
```typescript
// e2e-qa/features/goods-escrow.spec.ts
test('Pembeli konfirmasi penerimaan barang', async ({ page }) => {
  // Given
  await loginAs(page, 'budi@test.com')
  // When
  await page.goto('/transactions/TRX-001')
  await page.click('[data-testid="confirm-receipt"]')
  // Then
  await expect(page.locator('[data-testid="tx-status"]'))
    .toHaveText('COMPLETED')
})
```

### Link ke File

```
Scenario                          → Feature File                      → Test File
────────────────────────────────────────────────────────────────────────────────
Goods escrow happy path           → e2e-qa/features/goods.feature    → goods_test.go
Milestone payment                 → e2e-qa/features/services.feature → services_test.go
Wallet top-up                     → e2e-qa/features/wallet.feature   → wallet_test.go
KYC verification                  → e2e-qa/features/kyc.feature      → kyc_test.go
Dispute resolution                → e2e-qa/features/dispute.feature  → dispute_test.go
Vendor marketplace                → e2e-qa/features/vendor.feature   → vendor_test.go
```