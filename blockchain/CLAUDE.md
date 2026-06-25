# 🤖 Claude AI - Blockchain (Audit Log)

Panduan lokal ini menegaskan batasan ketat smart contract di ekosistem **RekberKuy**.

## ⛓️ BATASAN MUTLAK (Audit Log Only)
Sesuai regulasi di Indonesia yang belum melegalkan pembayaran kripto, sistem menerapkan arsitektur **Gasless Transaction**. Ini berarti pengguna *frontend/mobile* tidak akan berinteraksi dengan blockchain. Smart contract **HANYA** berfungsi sebagai *Audit Log Immutable* yang dipicu (*triggered*) di latar belakang oleh Backend sebagai *Relayer*.

**✅ YANG BOLEH DILAKUKAN:**
- Mencatat ID / Hash transaksi escrow yang sudah Selesai (Released/Refunded).
- Menyimpan timestamp (waktu pencatatan) dan status akhir.
- Mengeluarkan (emit) events untuk transparansi platform.

**❌ YANG DILARANG KERAS:**
- Menahan, menyimpan, atau memindahkan dana escrow (semuanya harus di handle Midtrans/RekberPay di Backend).
- Menulis logika kalkulasi fee, komisi, denda, atau logika bisnis apapun.
- Menyimpan data PII/Sensitif pengguna (termasuk dilarang keras menaruhnya di dalam tipe `mapping`).

## 🛠️ Stack & Network
- **Framework**: Hardhat v3.
- **Bahasa**: Solidity.
- **Network Target**: Avalanche C-Chain (Fuji Testnet & Mainnet).
- **Deployment Tool**: Hardhat Ignition (`ignition/modules/`).

## 💻 Command Penting
```bash
npm install
npx hardhat compile      # Kompilasi Smart Contract
npx hardhat test         # Jalankan Unit Test kontrak
npx hardhat node         # Local blockchain (Port 8545)
```