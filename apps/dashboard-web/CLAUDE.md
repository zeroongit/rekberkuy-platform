@AGENTS.md

# 🤖 Claude AI - Dashboard Web (Frontend)

Panduan lokal ini memberikan instruksi spesifik untuk pengembangan UI **RekberKuy** menggunakan Next.js.

## 💻 Stack & Aturan Utama
- **Framework**: Next.js v16 (Strictly **App Router** di `src/app/`, dilarang menggunakan Pages Router).
- **Styling**: Tailwind CSS v4 murni. DILARANG menggunakan CSS modules atau styled-components.
- **Komponen UI**: Wajib mendahulukan instalasi/pemakaian **Shadcn/UI** sebelum membuat komponen dari awal.

## 📐 Konvensi React & TypeScript
- **Server Components by Default**: Komponen berjalan di server secara default. Tambahkan direktif `"use client"` di baris paling atas HANYA jika membutuhkan interaktivitas (seperti `useState`, `useEffect`, `onClick`).
- **TypeScript Strict**: Hindari tipe `any` kecuali sangat terpaksa. Gunakan `interface` untuk mendefinisikan Props pada komponen React, dan `type` untuk union/intersection.
- **Penamaan File**: 
  - Komponen UI: `PascalCase.tsx` (contoh: `TransactionCard.tsx`)
  - Hooks / Utils: `camelCase.ts` (contoh: `useWallet.ts`, `formatCurrency.ts`)

## 🛠️ Command Penting
```bash
npm run dev    # Jalankan dev server lokal
npm run lint   # Cek linter
npm run test   # Jalankan Jest unit test
```
