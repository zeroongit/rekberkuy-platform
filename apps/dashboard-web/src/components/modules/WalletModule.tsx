'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Dialog, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/Dialog';
import { WalletTopupForm } from '@/components/forms/WalletTopupForm';
import { useWalletStore } from '@/store/useWalletStore';
import { notify } from '@/components/providers/ToastProvider';
import { Wallet, ArrowDownRight, ArrowUpRight, Shield, Zap } from 'lucide-react';

export function WalletModule() {
  const { wallet, updateBalance } = useWalletStore();
  const [isTopupOpen, setIsTopupOpen] = useState(false);
  const [isWithdrawOpen, setIsWithdrawOpen] = useState(false);

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
    notify.success('Saldo masuk Rp 500.000 via Real-time WebSocket Event!');
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Wallet Card */}
      <Card className="lg:col-span-1 border-blue-200 dark:border-blue-900/50 bg-gradient-to-br from-blue-600 to-indigo-700 text-white shadow-xl relative overflow-hidden">
        <div className="absolute top-0 right-0 -mt-6 -mr-6 w-32 h-32 bg-white/10 rounded-full blur-2xl pointer-events-none" />

        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <span className="flex items-center space-x-2 text-xs font-semibold uppercase tracking-wider text-blue-100">
              <Wallet className="h-4 w-4" />
              <span>RekberPay Wallet</span>
            </span>
            <span className="px-2.5 py-0.5 bg-white/20 backdrop-blur-md rounded-full text-xs font-bold">
              {wallet?.currency || 'IDR'}
            </span>
          </div>
          <CardTitle className="text-3xl font-black mt-2 text-white">{formattedBalance}</CardTitle>
          <CardDescription className="text-xs text-blue-100">
            Saldo aktif siap pakai & transaksi
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4 pt-2">
          <div className="p-3 bg-white/10 backdrop-blur-xs rounded-xl flex items-center justify-between text-xs">
            <span className="text-blue-100 flex items-center space-x-1.5">
              <Shield className="h-3.5 w-3.5" />
              <span>Dana Terkunci Escrow:</span>
            </span>
            <span className="font-extrabold text-white">{formattedLocked}</span>
          </div>

          <div className="grid grid-cols-2 gap-2 pt-2">
            <Button
              onClick={() => setIsTopupOpen(true)}
              className="w-full bg-white text-blue-700 hover:bg-blue-50 font-bold rounded-xl text-xs h-9 shadow-sm"
            >
              <ArrowDownRight className="h-4 w-4 mr-1" />
              Top-Up
            </Button>
            <Button
              onClick={() => setIsWithdrawOpen(true)}
              variant="outline"
              className="w-full border-white/30 text-white hover:bg-white/10 font-bold rounded-xl text-xs h-9"
            >
              <ArrowUpRight className="h-4 w-4 mr-1" />
              Tarik Dana
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Security & Overview Card */}
      <Card className="lg:col-span-2 flex flex-col justify-between p-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2.5">
              <div className="p-2 rounded-xl bg-blue-50 dark:bg-blue-950/50 text-blue-600 dark:text-blue-400">
                <Zap className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-zinc-900 dark:text-zinc-100">
                  Sistem Escrow & Real-time Sinkronisasi
                </h3>
                <p className="text-xs text-zinc-500">
                  Avalanche Blockchain Gasless Relayer + Redis Cache Layer
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleSimulateRealtimeDeposit}
              className="text-xs"
            >
              Simulasi Event
            </Button>
          </div>

          <p className="text-sm text-zinc-600 dark:text-zinc-400 leading-relaxed">
            Setiap dana yang masuk ke dalam sistem RekberPay langsung diamankan dalam smart contract audit log Avalanche secara otomatis tanpa biaya gas untuk pengguna.
          </p>
        </div>

        <div className="grid grid-cols-3 gap-4 pt-4 border-t border-zinc-100 dark:border-zinc-800 text-center">
          <div>
            <p className="text-xs text-zinc-500">Status Keamanan</p>
            <p className="text-xs font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">Tervalidasi Zod</p>
          </div>
          <div>
            <p className="text-xs text-zinc-500">Audit Relayer</p>
            <p className="text-xs font-bold text-blue-600 dark:text-blue-400 mt-0.5">Avalanche Fuji</p>
          </div>
          <div>
            <p className="text-xs text-zinc-500">Cache Latensi</p>
            <p className="text-xs font-bold text-zinc-800 dark:text-zinc-200 mt-0.5">&lt; 5ms (Redis)</p>
          </div>
        </div>
      </Card>

      {/* Top-up Dialog */}
      <Dialog open={isTopupOpen} onOpenChange={setIsTopupOpen}>
        <DialogHeader>
          <DialogTitle>Top-Up Saldo RekberPay</DialogTitle>
          <DialogDescription>
            Pilih nominal dan metode pembayaran melalui Midtrans payment gateway yang aman.
          </DialogDescription>
        </DialogHeader>
        <WalletTopupForm />
      </Dialog>

      {/* Withdraw Dialog Placeholder */}
      <Dialog open={isWithdrawOpen} onOpenChange={setIsWithdrawOpen}>
        <DialogHeader>
          <DialogTitle>Tarik Dana ke Rekening Bank</DialogTitle>
          <DialogDescription>
            Pencairan dana langsung ke rekening bank terdaftar Anda.
          </DialogDescription>
        </DialogHeader>
        <div className="p-4 bg-zinc-50 dark:bg-zinc-800/50 rounded-xl text-center space-y-3">
          <p className="text-sm text-zinc-600 dark:text-zinc-400">
            Fitur penarikan dana diproteksi dengan otorisasi 2-Faktor dan verifikasi KYC.
          </p>
          <Button onClick={() => setIsWithdrawOpen(false)} className="w-full text-xs">
            Tutup
          </Button>
        </div>
      </Dialog>
    </div>
  );
}
