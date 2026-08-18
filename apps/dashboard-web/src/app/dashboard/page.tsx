import React from 'react';
import { Navbar } from '@/components/modules/Navbar';
import { WalletModule } from '@/components/modules/WalletModule';

export default function DashboardOverviewRoute() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        <h1 className="text-2xl font-black text-zinc-900 dark:text-zinc-50">Ringkasan Dompet & Wallet</h1>
        <WalletModule />
      </main>
    </div>
  );
}
