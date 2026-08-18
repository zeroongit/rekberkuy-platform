'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { notify } from '@/components/providers/ToastProvider';
import { AlertTriangle, ShieldAlert, Send } from 'lucide-react';

export function DisputeModule() {
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmitDispute = (e: React.FormEvent) => {
    e.preventDefault();
    if (!reason || reason.length < 10) {
      notify.error('Alasan sengketa minimal 10 karakter.');
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      setIsSubmitting(false);
      notify.success('Pengajuan sengketa (Dispute) berhasil dikirim ke Admin RekberKuy.');
      setReason('');
    }, 800);
  };

  return (
    <Card className="p-6 border-red-200 dark:border-red-900/50">
      <CardHeader className="px-0 pt-0">
        <div className="flex items-center space-x-3">
          <div className="p-2 rounded-xl bg-red-50 dark:bg-red-950/50 text-red-600 dark:text-red-400">
            <ShieldAlert className="h-6 w-6" />
          </div>
          <div>
            <CardTitle className="text-xl">Panel Mediasi & Sengketa (Dispute)</CardTitle>
            <CardDescription className="text-xs mt-0.5">
              Pembekuan dana escrow sementara untuk mediasi admin netral
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0 pb-0 space-y-4">
        <div className="p-4 rounded-xl bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900 text-amber-800 dark:text-amber-200 text-xs flex items-start space-x-3">
          <AlertTriangle className="h-5 w-5 shrink-0 text-amber-600 mt-0.5" />
          <p>
            Mengajukan sengketa akan membekukan rilis dana secara otomatis hingga mediator Admin memeriksa bukti valid dari kedua belah pihak dan memberikan keputusan akhir (`REFUND_BUYER` atau `RELEASE_TO_SELLER`).
          </p>
        </div>

        <form onSubmit={handleSubmitDispute} className="space-y-4">
          <div>
            <label className="block text-xs font-bold text-zinc-700 dark:text-zinc-300 mb-1">
              Alasan / Kronologi Sengketa
            </label>
            <textarea
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Jelaskan kendala transaksi secara detail..."
              className="w-full rounded-xl border border-zinc-200 dark:border-zinc-800 bg-transparent p-3 text-sm text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <Button
            type="submit"
            variant="destructive"
            disabled={isSubmitting}
            className="w-full text-xs font-bold"
          >
            <Send className="h-4 w-4 mr-2" />
            {isSubmitting ? 'Mengirim...' : 'Kirim Laporan Sengketa'}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
