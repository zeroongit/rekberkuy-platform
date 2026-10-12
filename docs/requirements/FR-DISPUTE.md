# Functional Requirements: Sengketa & Mediasi (FR-DISPUTE)

## 1. Overview
Modul sengketa mengatur proses eskalasi perselisihan transaksi di mana dana escrow dibekukan (*frozen*), pihak-pihak melampirkan bukti, dan Admin bertindak sebagai mediator yang memberikan keputusan mengikat (`REFUND_BUYER` atau `RELEASE_TO_SELLER`). Sesuai [ADR-0003](../technical/adr/0003-admin-mediated-dispute-resolution-ai-deferred.md), AI tidak melakukan auto-resolve sengketa.

## 2. Requirement List

| ID | Requirement | Priority | Implementation Status |
|----|-------------|----------|------------------------|
| **FR-DISPUTE-01** | Salah satu pihak (buyer/seller/client) dapat membuka sengketa selama transaksi dalam status aktif. | M | Implemented (`dispute_usecase.go`) |
| **FR-DISPUTE-02** | Pembuka sengketa wajib mengisi alasan dan melampirkan bukti pendukung (foto, dokumen). | M | Implemented |
| **FR-DISPUTE-03** | Dana escrow secara otomatis dibekukan (`DISPUTED`) begitu sengketa dibuka. | M | Implemented |
| **FR-DISPUTE-04** | Pihak terlapor menerima notifikasi dan diberikan waktu 2x24 jam untuk memberikan tanggapan. | M | Implemented |
| **FR-DISPUTE-05** | Admin masuk sebagai mediator untuk meninjau bukti dari kedua belah pihak. | M | Implemented |
| **FR-DISPUTE-06** | Admin memutuskan outcome sengketa: `REFUND_BUYER` (dana dikembalikan ke pembeli) atau `RELEASE_TO_SELLER` (dana diberikan ke penjual). | M | Implemented |
| **FR-DISPUTE-07** | Keputusan admin bersifat final dan langsung dieksekusi secara otomatis oleh sistem (unit of work). | M | Implemented |
| **FR-DISPUTE-08** | SLA target penyelesaian sengketa adalah maksimal 5x24 jam. | M | Implemented |
| **FR-DISPUTE-09** | Seluruh catatan mediasi dan keputusan admin terekam dalam audit trail sistem. | M | Implemented |
