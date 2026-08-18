'use client';

import React, { useState, useMemo } from 'react';
import Link from 'next/link';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/Table';
import { EscrowStatusBadge } from '@/components/atoms/EscrowStatusBadge';
import { Transaction } from '@/types';
import { Search, ExternalLink } from 'lucide-react';

interface TransactionListModuleProps {
  initialTransactions?: Transaction[];
}

const DEFAULT_TRANSACTIONS: Transaction[] = [
  {
    id: 'f47ac10b-58cc-4372-a567-0e02b2c3d479',
    title: 'Pembelian Laptop Gaming ASUS ROG Strix G16',
    type: 'GOODS',
    status: 'FUNDS_LOCKED',
    amount: 18500000,
    buyerId: 'user-1',
    sellerId: 'merchant-1',
    buyerName: 'Budi Santoso',
    sellerName: 'TechStore ID',
    createdAt: '2026-08-18T10:00:00Z',
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440000',
    title: 'Jasa Pembuatan Fullstack Web App Rekber',
    type: 'SERVICES',
    status: 'WAITING_PAYMENT',
    amount: 7500000,
    buyerId: 'user-1',
    sellerId: 'provider-2',
    buyerName: 'Budi Santoso',
    sellerName: 'DevHouse Studio',
    createdAt: '2026-08-18T11:30:00Z',
  },
  {
    id: '6ba7b810-9dad-11d1-80b4-00c04fd430c8',
    title: 'Penyelenggaraan Wedding Outdoor & Sound System',
    type: 'EVENTS',
    status: 'RELEASED',
    amount: 45000000,
    buyerId: 'user-1',
    sellerId: 'vendor-3',
    buyerName: 'Budi Santoso',
    sellerName: 'Grand Wedding EO',
    createdAt: '2026-08-17T09:15:00Z',
    blockchainTxHash: '0xabc123456789abcdef123456789abcdef123456789abcdef123456789abcdef',
  },
];

export function TransactionListModule({
  initialTransactions = DEFAULT_TRANSACTIONS,
}: TransactionListModuleProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryTab, setCategoryTab] = useState('ALL');

  const filteredTransactions = useMemo(() => {
    return initialTransactions.filter((tx) => {
      const matchSearch =
        tx.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        tx.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        tx.sellerName?.toLowerCase().includes(searchQuery.toLowerCase());

      const matchCategory = categoryTab === 'ALL' || tx.type === categoryTab;

      return matchSearch && matchCategory;
    });
  }, [initialTransactions, searchQuery, categoryTab]);

  return (
    <Card className="p-6">
      <CardHeader className="px-0 pt-0">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <CardTitle className="text-xl">Daftar Transaksi Escrow</CardTitle>
            <CardDescription className="text-xs mt-1">
              Pantau status penguncian dana dan riwayat transaksi aman
            </CardDescription>
          </div>
          <Badge variant="outline" className="w-fit text-xs font-mono">
            {filteredTransactions.length} Total Transaksi
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="px-0 pb-0 space-y-4">
        {/* Controls: Category Filter + Search */}
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
          <Tabs
            defaultValue="ALL"
            value={categoryTab}
            onValueChange={setCategoryTab}
            className="w-full sm:w-auto"
          >
            <TabsList>
              <TabsTrigger value="ALL">Semua</TabsTrigger>
              <TabsTrigger value="GOODS">Barang</TabsTrigger>
              <TabsTrigger value="SERVICES">Jasa</TabsTrigger>
              <TabsTrigger value="EVENTS">Event (EO)</TabsTrigger>
            </TabsList>
          </Tabs>

          <div className="relative w-full sm:w-64">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-zinc-400" />
            <Input
              type="text"
              placeholder="Cari transaksi / UUID..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 text-xs"
            />
          </div>
        </div>

        {/* Transaction Table */}
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-800 overflow-hidden">
          <Table>
            <TableHeader className="bg-zinc-50 dark:bg-zinc-800/50">
              <TableRow>
                <TableHead className="w-[320px]">ID / Judul Transaksi</TableHead>
                <TableHead>Kategori</TableHead>
                <TableHead>Status Escrow</TableHead>
                <TableHead>Nominal</TableHead>
                <TableHead className="text-right">Aksi</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTransactions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-32 text-center text-zinc-500">
                    Tidak ada transaksi yang cocok dengan kriteria pencarian.
                  </TableCell>
                </TableRow>
              ) : (
                filteredTransactions.map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell>
                      <div className="space-y-0.5">
                        <span className="font-mono text-[11px] text-blue-600 dark:text-blue-400 font-bold block truncate max-w-[280px]">
                          {tx.id}
                        </span>
                        <span className="font-semibold text-zinc-900 dark:text-zinc-100 block">
                          {tx.title}
                        </span>
                        {tx.sellerName && (
                          <span className="text-[11px] text-zinc-500 block">
                            Pihak Kedua: {tx.sellerName}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary" className="text-[10px]">
                        {tx.type}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <EscrowStatusBadge status={tx.status} />
                    </TableCell>
                    <TableCell className="font-extrabold text-zinc-900 dark:text-zinc-100">
                      {new Intl.NumberFormat('id-ID', {
                        style: 'currency',
                        currency: 'IDR',
                        maximumFractionDigits: 0,
                      }).format(tx.amount)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Link href={`/dashboard/transactions/${tx.id}`}>
                        <Button size="sm" variant="default" className="text-xs h-8">
                          Detail
                          <ExternalLink className="h-3.5 w-3.5 ml-1.5" />
                        </Button>
                      </Link>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}
