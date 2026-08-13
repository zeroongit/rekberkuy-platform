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
   # Lokal (setelah `npx hardhat node` berjalan)
   npx hardhat ignition deploy ignition/modules/TransactionLogger.ts --network localhost

   # Fuji Testnet (butuh AVALANCHE_RPC_URL & DEPLOYER_PRIVATE_KEY di .env)
   npx hardhat ignition deploy ignition/modules/TransactionLogger.ts --network avalancheFuji
   ```
   > Setelah deploy, catat address kontrak lalu set `CONTRACT_ADDRESS` di
   > `apps/core-service/.env` agar Go Relayer (`internal/relayer/relayer.go`) bisa memanggilnya.

## 📜 Kontrak — `TransactionLogger.sol`

Satu-satunya kontrak production. Fungsi & event signature-nya **frozen** — harus byte-identical
dengan `loggerABI` di `apps/core-service/internal/relayer/relayer.go` (mengubahnya = breaking
change ke Go caller karena selector & event topic berubah):

```solidity
function logTransaction(bytes32 txIdHash, uint256 amount, bytes32 buyerHash, bytes32 sellerHash) external;
event   TransactionLogged(bytes32 indexed txIdHash, uint256 amount,
                          bytes32 indexed buyerHash, bytes32 indexed sellerHash, uint256 timestamp);
```

Karakteristik & batasan yang diterapkan:

- **Audit log only:** TIDAK pernah menahan/menyimpan uang, TIDAK menghitung komisi/fee, dan
  TIDAK menyimpan PII. Satu-satunya data on-chain adalah hash `bytes32` + `amount`, dan itu
  pun hanya sebagai **event log**. Satu-satunya state yang ditulis adalah slot `relayer` (lihat
  access control) — tidak ada ledger, balance, atau data user.
- **Access control (2 wallet):**
  - `owner` (`immutable`): wallet admin offline/cold. Hanya ini yang boleh memanggil
    `setRelayer`. Immutable agar tidak ter-cache di storage dan tidak bisa di-rotasi tanpa
    redeploy (sengaja — kunci admin sebaiknya hidup di cold storage).
  - `relayer` (`address public`, mutable): hot wallet backend yang menandatangani setiap
    `logTransaction`. Bisa di-rotasi oleh owner lewat `setRelayer` **tanpa redeploy**, jadi
    kunci relayer yang compromise bisa dicabut tanpa memecah audit trail ke kontrak baru.
  - `logTransaction` di-guard `onlyRelayer` (address lain `revert Unauthorized()`).
  - `setRelayer(address) external onlyOwner` → revert `NotOwner()` bila bukan owner, dan
    `ZeroRelayerAddress()` bila diberi `address(0)`.
  - Setiap perubahan relayer (termasuk assignment awal di constructor) emit
    `RelayerUpdated(address indexed oldRelayer, address indexed newRelayer)` agar seluruh
    riwayat rotasi tercatat publik di audit trail.
- **Constructor 2-arg:** `constructor(address _owner, address _relayer)`. Untuk deploy
  production, berikan dua wallet berbeda lewat `--parameters` (lihat
  `ignition/modules/TransactionLogger.ts`); default kedua-duanya ke Hardhat account 0 untuk
  dev/lokal.
- **Gas per call:** `logTransaction` membaca `relayer` sebagai **cold SLOAD** di modifier
  `onlyRelayer` (harga rotasi-mampu), lalu hanya `emit TransactionLogged`. Tidak ada SSTORE di
  path `logTransaction`. Estimasi terukur ~28.5k gas/call (bandingkan ~26.4k bila `relayer`
  immutable) — trade-off +~2.1k gas/call ditukar dengan kemampuan rotasi tanpa redeploy.
- **Gasless:** User frontend tidak butuh wallet. Backend Relayer (`DEPLOYER_PRIVATE_KEY`)
  menanggung semua gas. Karena dipanggil setiap transaksi escrow selesai (volume tinggi),
  biaya ini jadi biaya operasional platform — lihat estimasi gas dari `npx hardhat test`.


