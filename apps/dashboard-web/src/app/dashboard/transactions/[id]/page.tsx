import React from 'react';
import { notFound } from 'next/navigation';
import { uuidParamSchema } from '@/lib/validations/uuid.schema';
import { Navbar } from '@/components/modules/Navbar';
import { EscrowDetailModule } from '@/components/modules/EscrowDetailModule';
import { Transaction } from '@/types';

interface TransactionDetailPageProps {
  params: Promise<{ id: string }>;
}

async function fetchTransactionByUuid(id: string): Promise<Transaction | null> {
  // Mock data matching requested UUID v4
  return {
    id,
    title: 'Pembelian Laptop Gaming ASUS ROG Strix G16',
    type: 'GOODS',
    status: 'FUNDS_LOCKED',
    amount: 18500000,
    buyerId: 'f47ac10b-58cc-4372-a567-0e02b2c3d479',
    sellerId: '550e8400-e29b-41d4-a716-446655440000',
    buyerName: 'Budi Santoso',
    sellerName: 'TechStore Official ID',
    createdAt: new Date().toISOString(),
    blockchainTxHash: '0xabc123456789abcdef123456789abcdef123456789abcdef123456789abcdef',
    milestones: [
      { id: 'm-1', title: 'Verifikasi Barang & Pengiriman Kurir', amount: 18500000, status: 'PENDING' },
    ],
  };
}

export default async function TransactionDetailPage({ params }: TransactionDetailPageProps) {
  const resolvedParams = await params;

  // Strict Zod UUID validation to prevent IDOR & enumeration attacks
  const parseResult = uuidParamSchema.safeParse(resolvedParams);
  if (!parseResult.success) {
    notFound();
  }

  const { id } = parseResult.data;
  const transaction = await fetchTransactionByUuid(id);

  if (!transaction) {
    notFound();
  }

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8">
        <EscrowDetailModule transaction={transaction} />
      </main>

      <footer className="border-t border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 py-6 text-center text-xs text-zinc-500">
        <p>© 2026 RekberKuy Platform. Gasless Escrow Audit on Avalanche.</p>
      </footer>
    </div>
  );
}
