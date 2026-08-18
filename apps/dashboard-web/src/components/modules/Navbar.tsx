'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { DropdownMenu, DropdownMenuItem } from '@/components/ui/DropdownMenu';
import { Sheet, SheetHeader, SheetTitle } from '@/components/ui/Sheet';
import { ShieldCheck, Menu, Wallet, Bell, User, LogOut } from 'lucide-react';
import { useWalletStore } from '@/store/useWalletStore';
import { useNotificationStore } from '@/store/useNotificationStore';

export function Navbar() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { wallet } = useWalletStore();
  const { notifications } = useNotificationStore();

  const formattedBalance = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(wallet?.balance || 0);

  return (
    <header className="sticky top-0 z-40 w-full border-b border-zinc-200 dark:border-zinc-800 bg-white/80 dark:bg-zinc-900/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto flex h-16 items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand */}
        <div className="flex items-center space-x-3">
          <Link href="/" className="flex items-center space-x-2.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-md">
              <ShieldCheck className="h-6 w-6" />
            </div>
            <div>
              <span className="text-lg font-black tracking-tight text-zinc-900 dark:text-zinc-50">
                Rekber<span className="text-blue-600 dark:text-blue-400">Kuy</span>
              </span>
              <span className="hidden sm:inline-block ml-2 text-[10px] uppercase font-bold text-zinc-400">
                Escrow & EO
              </span>
            </div>
          </Link>
        </div>

        {/* Navigation Desktop */}
        <nav className="hidden lg:flex items-center space-x-4 text-xs font-medium text-zinc-600 dark:text-zinc-400">
          <Link href="/" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Beranda
          </Link>
          <Link href="/dashboard" className="text-blue-600 dark:text-blue-400 font-semibold transition-colors">
            Dashboard
          </Link>
          <Link href="/dashboard/transactions" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Transaksi
          </Link>
          <Link href="/dashboard/ledger" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Ledger
          </Link>
          <Link href="/dashboard/disputes" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Sengketa
          </Link>
          <Link href="/dashboard/catalog" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Marketplace
          </Link>
          <Link href="/dashboard/kyc" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            KYC
          </Link>
          <Link href="/dashboard/admin" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Admin
          </Link>
          <Link href="/auth" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
            Auth
          </Link>
        </nav>

        {/* Action Controls */}
        <div className="flex items-center space-x-3">
          {/* Wallet summary pill */}
          <div className="hidden sm:flex items-center space-x-2 px-3 py-1.5 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/50">
            <Wallet className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            <span className="text-xs font-bold text-zinc-900 dark:text-zinc-100">{formattedBalance}</span>
          </div>

          {/* Notification dropdown trigger */}
          <DropdownMenu
            trigger={
              <Button variant="ghost" size="icon" className="relative">
                <Bell className="h-5 w-5 text-zinc-600 dark:text-zinc-400" />
                {notifications.length > 0 && (
                  <span className="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-blue-600 ring-2 ring-white dark:ring-zinc-900" />
                )}
              </Button>
            }
          >
            <div className="p-3 w-72 text-sm">
              <div className="flex items-center justify-between mb-2">
                <span className="font-bold">Notifikasi</span>
                <Badge variant="secondary">{notifications.length}</Badge>
              </div>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {notifications.map((n) => (
                  <div key={n.id} className="p-2 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 text-xs">
                    <p className="font-semibold text-zinc-900 dark:text-zinc-100">{n.title}</p>
                    <p className="text-zinc-500 line-clamp-2 mt-0.5">{n.message}</p>
                  </div>
                ))}
              </div>
            </div>
          </DropdownMenu>

          {/* User profile dropdown */}
          <DropdownMenu
            trigger={
              <Button variant="ghost" size="icon" className="rounded-full bg-zinc-100 dark:bg-zinc-800">
                <User className="h-5 w-5 text-zinc-700 dark:text-zinc-300" />
              </Button>
            }
          >
            <div className="px-3 py-2 border-b border-zinc-100 dark:border-zinc-800">
              <p className="font-semibold text-xs text-zinc-900 dark:text-zinc-100">Budi Santoso</p>
              <p className="text-[10px] text-zinc-500">Verified User (KYC)</p>
            </div>
            <DropdownMenuItem className="text-xs">Profil Akun</DropdownMenuItem>
            <DropdownMenuItem className="text-xs">Keamanan & Password</DropdownMenuItem>
            <DropdownMenuItem className="text-xs text-red-600 dark:text-red-400 flex items-center space-x-2">
              <LogOut className="h-3.5 w-3.5" />
              <span>Keluar</span>
            </DropdownMenuItem>
          </DropdownMenu>

          {/* Mobile hamburger menu */}
          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={() => setMobileMenuOpen(true)}
          >
            <Menu className="h-5 w-5 text-zinc-700 dark:text-zinc-300" />
          </Button>
        </div>
      </div>

      {/* Mobile Drawer Navigation */}
      <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen} side="left">
        <SheetHeader>
          <SheetTitle>Menu Navigasi</SheetTitle>
        </SheetHeader>
        <div className="flex flex-col space-y-4 text-sm font-medium">
          <Link
            href="/"
            onClick={() => setMobileMenuOpen(false)}
            className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-blue-600 dark:text-blue-400 font-bold"
          >
            Dashboard
          </Link>
          <Link
            href="#transactions"
            onClick={() => setMobileMenuOpen(false)}
            className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
          >
            Transaksi Escrow
          </Link>
          <Link
            href="#marketplace"
            onClick={() => setMobileMenuOpen(false)}
            className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
          >
            Vendor Marketplace
          </Link>
        </div>
      </Sheet>
    </header>
  );
}
