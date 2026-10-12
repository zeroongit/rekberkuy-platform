'use client';

import React from 'react';
import { useWalletStore } from '@/store/useWalletStore';
import { notify } from '@/components/providers/ToastProvider';

export function WalletBalanceWidget() {
  const { wallet, updateBalance } = useWalletStore();

  const formattedBalance = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(wallet?.balance || 0);

  const formattedLocked = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(wallet?.lockedBalance || 0);

  const handleSimulateRealtimeDeposit = () => {
    const currentBalance = wallet?.balance || 0;
    const currentLocked = wallet?.lockedBalance || 0;
    const addedAmount = 500000;
    updateBalance(currentBalance + addedAmount, currentLocked);
    notify.success(`Dana masuk Rp 500.000 via Real-time WebSockets!`);
  };

  return (
    <div className="p-6 bg-gradient-to-br from-blue-600 to-indigo-700 text-white rounded-3xl shadow-lg relative overflow-hidden">
      <div className="absolute top-0 right-0 -mt-6 -mr-6 w-32 h-32 bg-white/10 rounded-full blur-2xl pointer-events-none" />

      <div className="flex items-center justify-between mb-4">
        <div>
          <p className="text-xs uppercase tracking-wider text-blue-100 font-medium">RekberPay Wallet</p>
          <h2 className="text-2xl font-extrabold mt-1">{formattedBalance}</h2>
        </div>
        <span className="px-3 py-1 bg-white/20 backdrop-blur-md rounded-full text-xs font-semibold">
          {wallet?.currency || 'IDR'}
        </span>
      </div>

      <div className="flex items-center justify-between pt-4 border-t border-white/20 text-xs text-blue-100">
        <div>
          <span>Dana Terkunci Escrow: </span>
          <strong className="text-white font-bold">{formattedLocked}</strong>
        </div>
        <button
          onClick={handleSimulateRealtimeDeposit}
          className="px-3 py-1.5 bg-white text-blue-700 hover:bg-blue-50 font-bold rounded-xl transition shadow-sm text-xs"
        >
          Simulasi Real-time
        </button>
      </div>
    </div>
  );
}
