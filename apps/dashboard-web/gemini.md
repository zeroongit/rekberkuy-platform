# 🧠 Gemini Code Assist - Dashboard Web (Frontend)

Dokumentasi lokal ini memberikan pedoman pengembangan antarmuka web platform **RekberKuy**.

## 💻 Tech Stack
- **Framework**: Next.js v16 (Strictly menggunakan **App Router** di `src/app/`, BUKAN Pages Router).
- **Styling**: Tailwind CSS v4 (Hanya Tailwind, dilarang menggunakan CSS module atau styled-components).
- **Komponen UI**: Shadcn/UI (berbasis Radix UI).
- **Bahasa**: TypeScript (semua file harus menggunakan `.ts` atau `.tsx`).

## 📐 Konvensi Pengembangan
1. **Server Components by Default**: Next.js App Router menggunakan Server Components secara default. Tambahkan direktif `"use client"` di baris paling atas HANYA jika komponen membutuhkan interaktivitas (misal: `useState`, `onClick`, `useEffect`, atau hooks lainnya).
2. **Shadcn/UI First**: Jika membutuhkan komponen standar (Button, Input, Dialog, Table), selalu gunakan atau install dari Shadcn/UI terlebih dahulu (`npx shadcn-ui@latest add <component>`). Jangan membuat komponen dasar dari awal kecuali tidak tersedia.
3. **Type Safety**:
   - Hindari penggunaan tipe `any` kecuali benar-benar tidak ada pilihan lain.
   - Gunakan `interface` untuk mendefinisikan Props pada komponen React.
   - Gunakan `type` untuk union atau intersection.
4. **Penamaan File**: 
   - Komponen React: `PascalCase.tsx` (contoh: `TransactionCard.tsx`).
   - Hooks atau Utils: `camelCase.ts` (contoh: `useWallet.ts`, `formatCurrency.ts`).

## 📂 Struktur Penting
- `src/app/`: Root directory untuk routing aplikasi (layout, pages, api routes jika ada).
- `public/`: Aset statis (gambar, SVG).

## 🛠️ Perintah Berguna
```bash
# Pindah ke direktori
cd apps/dashboard-web

# Install dependencies
npm install

# Menjalankan local development server
npm run dev

# Menjalankan Linter
npm run lint

# Build Production
npm run build

# Menjalankan Unit Test (Jest)
npm run test
```