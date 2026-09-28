'use client';

import React from 'react';
import { Button } from '@/components/ui/Button';
import { AlertCircle } from 'lucide-react';

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[50vh] gap-4 p-6">
      <div className="p-3 bg-red-50 dark:bg-red-950 text-red-600 dark:text-red-400 rounded-full">
        <AlertCircle className="w-8 h-8" />
      </div>
      <div className="text-center space-y-1">
        <h2 className="text-lg font-bold text-zinc-900 dark:text-zinc-50">Gagal Memuat Dashboard</h2>
        <p className="text-sm text-zinc-500 dark:text-zinc-400 max-w-md">
          {error.message || 'Terjadi kesalahan saat mengambil data dashboard.'}
        </p>
      </div>
      <Button onClick={reset} size="sm" className="mt-2">
        Coba Lagi
      </Button>
    </div>
  );
}
