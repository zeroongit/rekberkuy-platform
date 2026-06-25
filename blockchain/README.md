# ⛓️ RekberKuy - Blockchain (Audit Log)

Modul ini mengelola infrastruktur *smart contract* RekberKuy di jaringan **Avalanche**. 

## ⚠️ BATASAN MUTLAK (Audit Log Only)
Sesuai regulasi kripto di Indonesia, platform ini menerapkan skema **Gasless Transaction**. Artinya:
- **Fungsi Utama:** Hanya digunakan sebagai *Audit Log Immutable* untuk transparansi transaksi escrow.
- **Tanpa Gas Fee User:** Pengguna di *frontend* tidak menggunakan dompet kripto. Semua transaksi di-*broadcast* oleh Backend RekberKuy (sebagai *Relayer*) secara transparan.
- **DILARANG:** Menyimpan atau menahan uang/escrow, menghitung komisi/fee, atau menyimpan data PII pengguna di dalam *smart contract* (termasuk di dalam tipe *mapping*).

## 🛠️ Tech Stack & Jaringan
- **Framework:** Hardhat v3.
- **Deployment Tool:** Hardhat Ignition.
- **Bahasa:** Solidity.
- **Jaringan Target:** Avalanche C-Chain (Fuji Testnet untuk *dev/staging*, Mainnet untuk *prod*).

## 🚀 Cara Menjalankan & Deployment

1. **Install Dependensi:**
   ```bash
   npm install
   ```
2. **Kompilasi Kontrak:**
   ```bash
   npx hardhat compile
   ```
3. **Jalankan Test:**
   ```bash
   npx hardhat test
   ```
4. **Jalankan Node Lokal (Simulasi Blockchain):**
   ```bash
   npx hardhat node
   ```
5. **Deploy Kontrak (Hardhat Ignition):**
   ```bash
   npx hardhat ignition deploy ignition/modules/Counter.ts --network localhost
   ```

