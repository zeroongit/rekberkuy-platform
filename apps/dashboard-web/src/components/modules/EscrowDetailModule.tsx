'use client';

import React from 'react';
import Link from 'next/link';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { EscrowStatusBadge } from '@/components/atoms/EscrowStatusBadge';
import { Transaction } from '@/types';
import { ShieldCheck, User, ArrowLeft, CheckCircle2, AlertCircle } from 'lucide-react';
import { notify } from '@/components/providers/ToastProvider';

interface EscrowDetailModuleProps {
  transaction: Transaction;
}

export function EscrowDetailModule({ transaction }: EscrowDetailModuleProps) {
  const formattedAmount = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(transaction.amount);

  const handleReleaseFunds = () => {
    notify.success('Permintaan pelepasan dana berhasil dikirim ke Relayer Avalanche!');
  };

  const handleOpenDispute = () => {
    notify.warning('Form pengajuan dispute dibuka untuk mediasi Admin.');
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      {/* Navigation & Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <Link href="/">
          <Button variant="ghost" size="sm" className="text-xs">
            <ArrowLeft className="h-4 w-4 mr-1" />
            Kembali ke Dashboard
          </Button>
        </Link>
        <div className="flex items-center space-x-2">
          <EscrowStatusBadge status={transaction.status} />
          <Badge variant="outline" className="text-xs">
            {transaction.type}
          </Badge>
        </div>
      </div>

      {/* Main Details Card */}
      <Card className="p-6">
        <CardHeader className="px-0 pt-0 pb-4 border-b border-zinc-100 dark:border-zinc-800">
          <div className="flex items-start justify-between">
            <div>
              <span className="text-xs font-mono text-blue-600 dark:text-blue-400 font-bold block mb-1">
                ID Transaksi: {transaction.id}
              </span>
              <CardTitle className="text-2xl font-black">{transaction.title}</CardTitle>
            </div>
            <div className="text-right">
              <span className="text-xs text-zinc-500 block">Total Dana Escrow</span>
              <span className="text-2xl font-black text-blue-600 dark:text-blue-400">
                {formattedAmount}
              </span>
            </div>
          </div>
        </CardHeader>

        <CardContent className="px-0 pb-0 pt-6 space-y-6">
          {/* Party Cards */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="p-4 rounded-xl bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-800 space-y-2">
              <div className="flex items-center space-x-2 text-xs font-semibold text-zinc-500">
                <User className="h-4 w-4 text-blue-600" />
                <span>Pihak Pembeli (Buyer)</span>
              </div>
              <p className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                {transaction.buyerName || 'Budi Santoso'}
              </p>
              <p className="font-mono text-[11px] text-zinc-400 truncate">{transaction.buyerId}</p>
            </div>

            <div className="p-4 rounded-xl bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-800 space-y-2">
              <div className="flex items-center space-x-2 text-xs font-semibold text-zinc-500">
                <User className="h-4 w-4 text-emerald-600" />
                <span>Pihak Penjual / Penyedia Jasa</span>
              </div>
              <p className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                {transaction.sellerName || 'TechStore Official ID'}
              </p>
              <p className="font-mono text-[11px] text-zinc-400 truncate">{transaction.sellerId}</p>
            </div>
          </div>

          {/* Milestones Section if available */}
          {transaction.milestones && transaction.milestones.length > 0 && (
            <div className="space-y-3">
              <h4 className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
                Tahapan / Milestones Pembayaran
              </h4>
              <div className="space-y-2">
                {transaction.milestones.map((m, idx) => (
                  <div
                    key={m.id || idx}
                    className="flex items-center justify-between p-3 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900"
                  >
                    <div className="flex items-center space-x-3">
                      <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                      <span className="text-sm font-medium">{m.title}</span>
                    </div>
                    <span className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
                      {new Intl.NumberFormat('id-ID', {
                        style: 'currency',
                        currency: 'IDR',
                        maximumFractionDigits: 0,
                      }).format(m.amount)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Blockchain Audit Log */}
          {transaction.blockchainTxHash && (
            <div className="p-4 rounded-xl bg-blue-50/50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900 space-y-2">
              <div className="flex items-center space-x-2 text-xs font-bold text-blue-700 dark:text-blue-300">
                <ShieldCheck className="h-4 w-4" />
                <span>Catatan Audit Avalanche C-Chain (Gasless Log)</span>
              </div>
              <p className="font-mono text-xs text-blue-900 dark:text-blue-200 break-all">
                {transaction.blockchainTxHash}
              </p>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex flex-wrap items-center justify-end gap-3 pt-4 border-t border-zinc-100 dark:border-zinc-800">
            <Button variant="outline" onClick={handleOpenDispute} className="text-xs text-red-600 dark:text-red-400 border-red-200 dark:border-red-900">
              <AlertCircle className="h-4 w-4 mr-1.5" />
              Ajukan Sengketa (Dispute)
            </Button>
            <Button variant="default" onClick={handleReleaseFunds} className="text-xs">
              <CheckCircle2 className="h-4 w-4 mr-1.5" />
              Konfirmasi & Rilis Dana Escrow
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
