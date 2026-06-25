# 🧠 Gemini Code Assist - Blockchain (Audit Log)

Dokumentasi lokal ini memberikan pedoman mengenai batasan, fungsi, dan interaksi sistem Smart Contract di **RekberKuy**.

## ⛓️ Tujuan Utama & Batasan Mendasar
**SANGAT PENTING:** Modul blockchain ini (Smart Contract) berfungsi **MURNI SEBAGAI AUDIT LOG**. 

### ✅ Yang BOLEH dilakukan Smart Contract:
- Mencatat Hash / ID transaksi yang sudah *Selesai* (Released/Refunded).
- Mencatat Timestamp (waktu pencatatan on-chain).
- Mencatat status akhir dari transaksi.
- Mengeluarkan (`emit`) events untuk keperluan indexing, pencarian, dan transparansi data.

### ❌ Yang TIDAK BOLEH dilakukan Smart Contract:
- **DILARANG** menyimpan, menahan, atau memindahkan dana escrow secara on-chain.
- **DILARANG** meletakkan logika kalkulasi pembagian dana, potongan biaya layanan (fee), atau denda.
- **DILARANG** menyimpan data identitas sensitif pengguna (PII).
- **DILARANG** menggunakan `mapping` yang menyimpan data sensitif.

Seluruh logika uang dan bisnis dijalankan di backend (`apps/core-service/`). Mengingat pembayaran kripto belum legal di Indonesia, platform ini menggunakan arsitektur **Gasless Transaction**. Artinya, blockchain sepenuhnya beroperasi di latar belakang (background) dan dieksekusi oleh Backend (sebagai *relayer* pembayar gas fee). Pengguna akhir tidak mengetahui, tidak membutuhkan dompet kripto, dan tidak dibebani gas fee sama sekali.

## 💻 Tech Stack & Network
- **Framework**: Hardhat v3.
- **Bahasa**: Solidity.
- **Jaringan (Target)**: Avalanche C-Chain (Fuji Testnet untuk development, Mainnet untuk production).
- **Deployment**: Menggunakan Hardhat Ignition (`ignition/modules/`).

## 🛠️ Perintah Berguna
```bash
# Pindah ke direktori
cd blockchain

# Install dependencies
npm install

# Kompilasi Smart Contract
npx hardhat compile

# Menjalankan Test bawaan (Node Test Runner / Viem)
npx hardhat test

# Deploy ke Local Network (Jalankan `npx hardhat node` di tab lain)
npx hardhat ignition deploy ./ignition/modules/Counter.ts --network localhost
```