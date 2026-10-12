import React from 'react';
import { EscrowStatus } from '@/types';

interface EscrowStatusBadgeProps {
  status: EscrowStatus;
}

export function EscrowStatusBadge({ status }: EscrowStatusBadgeProps) {
  const statusConfig: Record<EscrowStatus, { label: string; className: string }> = {
    WAITING_PAYMENT: {
      label: 'Menunggu Pembayaran',
      className: 'bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-300 border-amber-200 dark:border-amber-900',
    },
    FUNDS_LOCKED: {
      label: 'Dana Dikunci (Escrow)',
      className: 'bg-blue-50 dark:bg-blue-950/40 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-900',
    },
    RELEASED: {
      label: 'Dana Dirilis',
      className: 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-900',
    },
    DISPUTED: {
      label: 'Dispute / Sengketa',
      className: 'bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-300 border-red-200 dark:border-red-900',
    },
    REFUNDED: {
      label: 'Dana Dikembalikan',
      className: 'bg-purple-50 dark:bg-purple-950/40 text-purple-700 dark:text-purple-300 border-purple-200 dark:border-purple-900',
    },
  };

  const config = statusConfig[status] || {
    label: status,
    className: 'bg-zinc-100 text-zinc-800 border-zinc-300',
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold border ${config.className}`}
    >
      {config.label}
    </span>
  );
}
