import React from 'react';
import Link from 'next/link';
import { Transaction } from '@/types';
import { EscrowStatusBadge } from '@/components/atoms/EscrowStatusBadge';

interface TransactionCardProps {
  transaction: Transaction;
  onActionClick?: (tx: Transaction) => void;
}

export function TransactionCard({ transaction, onActionClick }: TransactionCardProps) {
  const formattedAmount = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(transaction.amount);

  return (
    <div className="p-5 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-sm hover:shadow-md transition">
      <div className="flex items-start justify-between mb-3">
        <div>
          <span className="text-xs font-medium text-zinc-500 uppercase tracking-wide">
            {transaction.type}
          </span>
          <h3 className="text-base font-bold text-zinc-900 dark:text-zinc-100">
            {transaction.title}
          </h3>
        </div>
        <EscrowStatusBadge status={transaction.status} />
      </div>

      <div className="flex items-center justify-between mt-4 pt-3 border-t border-zinc-100 dark:border-zinc-800">
        <div>
          <p className="text-xs text-zinc-500">Nominal Escrow</p>
          <p className="text-lg font-extrabold text-blue-600 dark:text-blue-400">
            {formattedAmount}
          </p>
        </div>

        <div className="flex space-x-2">
          <Link
            href={`/dashboard/transactions/${transaction.id}`}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl transition shadow-sm"
          >
            Lihat Detail
          </Link>
          {onActionClick && (
            <button
              onClick={() => onActionClick(transaction)}
              className="px-3 py-2 bg-zinc-200 hover:bg-zinc-300 dark:bg-zinc-800 dark:hover:bg-zinc-700 text-zinc-800 dark:text-zinc-200 text-xs font-semibold rounded-xl transition"
            >
              Cepat
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
