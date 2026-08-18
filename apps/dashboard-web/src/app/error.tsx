'use client';

import React, { useEffect } from 'react';
import Link from 'next/link';

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log error to monitoring service in production
    console.error('[Dashboard Error Boundary]:', error);
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 p-6">
      <div className="max-w-md w-full bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-xl p-8 text-center">
        <div className="w-16 h-16 bg-red-100 dark:bg-red-950 text-red-600 dark:text-red-400 rounded-full flex items-center justify-center mx-auto mb-4 text-2xl">
          ⚠️
        </div>
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100 mb-2">
          Terjadi Kesalahan Sistem
        </h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-6">
          Maaf, terjadi kendala saat memproses halaman ini. Tim kami telah diberi tahu terkait masalah ini.
        </p>
        <div className="flex space-x-3 justify-center">
          <button
            onClick={() => reset()}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl text-sm transition shadow-md"
          >
            Coba Lagi
          </button>
          <Link
            href="/"
            className="px-4 py-2 bg-zinc-200 hover:bg-zinc-300 dark:bg-zinc-800 dark:hover:bg-zinc-700 text-zinc-800 dark:text-zinc-200 font-medium rounded-xl text-sm transition"
          >
            Kembali ke Beranda
          </Link>
        </div>
      </div>
    </div>
  );
}
