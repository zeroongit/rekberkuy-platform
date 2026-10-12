# Functional Requirements: Transaksi Barang (FR-TRANSACTION-GOODS)

## 1. Overview
Modul transaksi barang mengatur alur escrow jual beli fisik maupun digital, mulai dari pembuatan transaksi, penguncian dana, verifikasi fraud oleh AI, konfirmasi pengiriman resi, hingga rilis dana dan pencatatan audit on-chain.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-TRX-GOODS-01** | Pembeli dapat menginisiasi transaksi barang dengan mengisi detail produk, harga, dan memilih penjual terverifikasi. | M | Implemented (`transaction_goods_usecase.go`) |
| **FR-TRX-GOODS-02** | Sistem mendebet saldo pembeli dan menempatkan dana ke rekening escrow (`FUNDS_LOCKED`). | M | Implemented |
| **FR-TRX-GOODS-03** | AI fraud scoring (`backend-ai`) dieksekusi untuk mengevaluasi risiko sebelum dana dikunci ke escrow. | M | Implemented (`fraud/client.go`) |
| **FR-TRX-GOODS-04** | Penjual menerima notifikasi instan ketika transaksi baru dibuat dan dana dikunci. | M | Implemented |
| **FR-TRX-GOODS-05** | Penjual mengonfirmasi pesanan dan memasukkan nomor resi pengiriman kurir. | M | Implemented |
| **FR-TRX-GOODS-06** | Pembeli mengonfirmasi penerimaan barang setelah paket tiba. | M | Implemented |
| **FR-TRX-GOODS-07** | Sistem merilis dana dari escrow ke dompet penjual (dikurangi fee platform) setelah konfirmasi pembeli. | M | Implemented |
| **FR-TRX-GOODS-08** | Auto-release worker otomatis merilis dana ke penjual jika pembeli tidak merespons dalam 3 hari setelah estimasi tiba. | M | Implemented (`auto_release_worker.go`) |
| **FR-TRX-GOODS-09** | Transaksi yang selesai otomatis mencatatkan audit log ke blockchain Avalanche via gasless relayer. | M | Implemented (`relayer.go`) |
| **FR-TRX-GOODS-10** | Platform memotong komisi sesuai tier CRM penjual pada saat rilis dana. | M | Implemented (`finance_calculator.go`) |
