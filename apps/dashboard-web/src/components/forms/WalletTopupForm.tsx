'use client';

import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { topupWalletSchema, TopupWalletInput } from '@/lib/validations/wallet.schema';
import { FormErrorAlert } from '@/components/ui/FormErrorAlert';
import { FieldError } from '@/components/ui/FieldError';
import { useServerAction } from '@/hooks/useServerAction';
import { topupWalletAction } from '@/actions/wallet.actions';

export function WalletTopupForm() {
  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<TopupWalletInput>({
    resolver: zodResolver(topupWalletSchema),
    defaultValues: {
      amount: 50000,
      paymentMethod: 'VIRTUAL_ACCOUNT',
    },
  });

  const { execute, isPending, error, success, data } = useServerAction(
    topupWalletSchema,
    topupWalletAction
  );

  const onSubmit = async (values: TopupWalletInput) => {
    const res = await execute(values);
    if (res.success) {
      reset();
    }
  };

  return (
    <div className="max-w-md mx-auto p-6 bg-white dark:bg-zinc-900 rounded-xl shadow-md border border-zinc-200 dark:border-zinc-800">
      <h2 className="text-xl font-bold mb-4 text-zinc-900 dark:text-zinc-100">Top-Up RekberPay Wallet</h2>
      
      <FormErrorAlert message={error} />

      {success && data && (
        <div className="mb-4 p-3 bg-green-50 dark:bg-green-950/50 border border-green-200 dark:border-green-900 text-green-700 dark:text-green-300 rounded-lg text-sm">
          Top-up berhasil diinisiasi! Order ID: <strong>{(data as { orderId: string }).orderId}</strong>
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Jumlah Top-Up (IDR)
          </label>
          <input
            type="number"
            {...register('amount', { valueAsNumber: true })}
            className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-lg bg-transparent text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <FieldError message={errors.amount?.message} />
        </div>

        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Metode Pembayaran
          </label>
          <select
            {...register('paymentMethod')}
            className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded-lg bg-transparent text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="VIRTUAL_ACCOUNT">Virtual Account (VA)</option>
            <option value="QRIS">QRIS</option>
            <option value="BANK_TRANSFER">Bank Transfer</option>
          </select>
          <FieldError message={errors.paymentMethod?.message} />
        </div>

        <button
          type="submit"
          disabled={isPending}
          className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition disabled:opacity-50"
        >
          {isPending ? 'Memproses...' : 'Proses Top-Up'}
        </button>
      </form>
    </div>
  );
}
