# 💻 RekberKuy - Dashboard Web (Frontend)

Modul ini adalah antarmuka web modern (Frontend) untuk pengguna dan admin **RekberKuy**, dibangun menggunakan ekosistem **Next.js v16**.

## 🛠️ Tech Stack & Standar Pengembangan

- **Framework:** Next.js v16 (Strictly **App Router** di `src/app/`).
- **Styling:** Tailwind CSS v4. *Dilarang keras menggunakan CSS modules atau styled-components.*
- **Komponen UI:** Shadcn/UI. Prioritaskan pemakaian komponen bawaan dari library ini sebelum membuat dari awal.
- **Bahasa:** TypeScript. Tipe `any` sangat tidak disarankan kecuali dalam kondisi yang memaksa.

> **Catatan Penting:** Sebagian besar komponen Next.js di proyek ini berjalan sebagai **Server Components** secara *default*. Hanya gunakan direktif `"use client"` jika komponen benar-benar membutuhkan state atau hook React (seperti `useState`, `onClick`, `useEffect`).

## 🚀 Cara Menjalankan Lokal

1. **Install Dependensi:**
   ```bash
   npm install
   ```
2. **Jalankan Server Development:**
   ```bash
   npm run dev
   ```
   Aplikasi akan berjalan di http://localhost:3000.

3. **Build untuk Produksi:**
   ```bash
   npm run build
   ```

## 🧪 Pengujian & Linter
Setiap PR wajib lolos linter dan unit test.
```bash
# Cek linter
npm run lint
 
# Unit test (Jest)
npm run test
```
