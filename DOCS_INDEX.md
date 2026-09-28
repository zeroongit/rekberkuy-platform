# 📚 RekberKuy — Documentation Index

Sebelum mengerjakan task apapun, baca file ini untuk tahu
dokumen mana yang perlu dibaca berdasarkan konteks task.

## Wajib dibaca selalu
- `CLAUDE.md` — konvensi kode, arsitektur, aturan layer
- `docs/BRD.md` — kebutuhan bisnis & business rules

## Baca sesuai domain task

| Jika task menyentuh... | Baca ini |
|------------------------|----------|
| Auth, login, JWT, session | `docs/requirements/FR-AUTH.md` |
| KYC, verifikasi identitas | `docs/requirements/FR-KYC.md` |
| Wallet, top-up, withdraw | `docs/requirements/FR-WALLET.md` |
| Transaksi barang | `docs/requirements/FR-TRANSACTION-GOODS.md` |
| Transaksi jasa, milestone | `docs/requirements/FR-TRANSACTION-SERVICES.md` |
| Event, tiket, vendor | `docs/requirements/FR-TRANSACTION-EVENTS.md` |
| Sengketa, mediasi | `docs/requirements/FR-DISPUTE.md` |
| Vendor marketplace | `docs/requirements/FR-VENDOR.md` |
| Blockchain, relayer | `docs/requirements/FR-BLOCKCHAIN.md` |
| Alur bisnis & state machine | `docs/application-flow-and-module-guide.md` |
| Arsitektur teknis | `docs/technical/architecture.md` |
| Deployment, infra | `docs/deployment-vps-docker-compose.md` |