'use client';

import React from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/Table';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';

export interface LedgerEntry {
  id: string;
  type: 'TOPUP' | 'ESCROW_LOCK' | 'RELEASE' | 'REFUND' | 'WITHDRAWAL';
  amount: number;
  description: string;
  createdAt: string;
}

const MOCK_LEDGER: LedgerEntry[] = [
  {
    id: 'LEDGER-101',
    type: 'TOPUP',
    amount: 5000000,
    description: 'Top-up saldo RekberPay via Virtual Account BCA',
    createdAt: '2026-08-18T09:00:00Z',
  },
  {
    id: 'LEDGER-102',
    type: 'ESCROW_LOCK',
    amount: -18500000,
    description: 'Penguncian dana escrow untuk transaksi #TRX-1001',
    createdAt: '2026-08-18T10:00:00Z',
  },
  {
    id: 'LEDGER-103',
    type: 'RELEASE',
    amount: 45000000,
    description: 'Penerimaan rilis dana escrow dari transaksi Event EO',
    createdAt: '2026-08-17T15:30:00Z',
  },
];

export function WalletLedgerModule() {
  return (
    <Card className="p-6">
      <CardHeader className="px-0 pt-0">
        <CardTitle className="text-xl">Buku Besar RekberPay (Ledger)</CardTitle>
        <CardDescription className="text-xs mt-1">
          Catatan mutasi saldo, top-up, rilis, pengembalian, dan pencairan dana
        </CardDescription>
      </CardHeader>
      <CardContent className="px-0 pb-0">
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-800 overflow-hidden">
          <Table>
            <TableHeader className="bg-zinc-50 dark:bg-zinc-800/50">
              <TableRow>
                <TableHead>ID Mutasi</TableHead>
                <TableHead>Tipe</TableHead>
                <TableHead>Keterangan</TableHead>
                <TableHead>Nominal</TableHead>
                <TableHead className="text-right">Waktu</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {MOCK_LEDGER.map((entry) => {
                const isPositive = entry.amount > 0;
                return (
                  <TableRow key={entry.id}>
                    <TableCell className="font-mono text-xs font-bold text-blue-600 dark:text-blue-400">
                      {entry.id}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={isPositive ? 'success' : 'secondary'}
                        className="text-[10px]"
                      >
                        {entry.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {entry.description}
                    </TableCell>
                    <TableCell>
                      <span
                        className={`font-extrabold flex items-center space-x-1 ${
                          isPositive ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'
                        }`}
                      >
                        {isPositive ? <ArrowDownRight className="h-4 w-4 mr-1" /> : <ArrowUpRight className="h-4 w-4 mr-1" />}
                        {new Intl.NumberFormat('id-ID', {
                          style: 'currency',
                          currency: 'IDR',
                          maximumFractionDigits: 0,
                        }).format(entry.amount)}
                      </span>
                    </TableCell>
                    <TableCell className="text-right text-xs text-zinc-500">
                      {new Date(entry.createdAt).toLocaleString('id-ID')}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
