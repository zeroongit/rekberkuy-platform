import React from 'react';
import { Navbar } from '@/components/modules/Navbar';
import { TransactionListModule } from '@/components/modules/TransactionListModule';

export default function TransactionsPageRoute() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        <TransactionListModule />
      </main>
    </div>
  );
}
