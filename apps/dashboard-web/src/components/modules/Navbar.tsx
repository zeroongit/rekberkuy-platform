'use client';

import React, { useState, useSyncExternalStore } from 'react';
import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { DropdownMenu, DropdownMenuItem } from '@/components/ui/DropdownMenu';
import { Sheet, SheetHeader, SheetTitle } from '@/components/ui/Sheet';
import { ShieldCheck, Menu, Wallet, Bell, LogOut } from 'lucide-react';
import { useWalletStore } from '@/store/useWalletStore';
import { useNotificationStore } from '@/store/useNotificationStore';
import { useAuthStore } from '@/store/useAuthStore';
import { logoutAction } from '@/actions/auth.actions';
import { notify } from '@/components/providers/ToastProvider';

export function Navbar() {
  const mounted = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false
  );
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { wallet } = useWalletStore();
  const { notifications } = useNotificationStore();
  const { user, isAuthenticated, logout } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();

  const handleLogout = async () => {
    try {
      await logoutAction();
      logout();
      notify.success('Berhasil keluar dari akun.');
      router.push('/auth');
    } catch {
      notify.error('Gagal keluar.');
    }
  };

  const formattedBalance = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(wallet?.balance || 0);

  const isAdmin = user?.role === 'ADMIN';

  return (
    <header className="sticky top-0 z-40 w-full border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 shadow-xs">
      <div className="max-w-7xl mx-auto flex h-16 items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand / Logo */}
        <div className="flex items-center space-x-4">
          <Link href={isAdmin ? "/dashboard/admin" : (isAuthenticated ? "/dashboard" : "/")} className="flex items-center space-x-2.5">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-md">
              <ShieldCheck className="h-6 w-6" />
            </div>
            <div>
              <span className="text-lg font-black tracking-tight text-zinc-900 dark:text-zinc-50">
                Rekber<span className="text-blue-600 dark:text-blue-400">Kuy</span>
              </span>
              <span className="hidden sm:inline-block ml-2 px-1.5 py-0.5 rounded text-[9px] uppercase font-bold bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400">
                {isAdmin ? 'Admin Panel' : 'Escrow & EO'}
              </span>
            </div>
          </Link>
        </div>

        {/* Navigation Desktop */}
        <nav className="hidden lg:flex items-center space-x-1 text-xs font-semibold text-zinc-600 dark:text-zinc-300">
          {mounted && isAuthenticated ? (
            isAdmin ? (
              <>
                <Link href="/dashboard/admin" className={`px-3 py-2 rounded-lg hover:bg-purple-50 dark:hover:bg-purple-950/30 text-purple-600 dark:text-purple-400 transition-colors ${pathname === '/dashboard/admin' ? 'bg-purple-50 dark:bg-purple-950/40 font-bold' : ''}`}>
                  Panel Admin
                </Link>
                <Link href="/dashboard/disputes" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/disputes' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Moderasi Sengketa
                </Link>
                <Link href="/dashboard/kyc" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/kyc' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Persetujuan KYC
                </Link>
                <Link href="/dashboard/transactions" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname.startsWith('/dashboard/transactions') ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Audit Pesanan
                </Link>
              </>
            ) : (
              <>
                <Link href="/dashboard/catalog" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/catalog' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Beranda (Katalog)
                </Link>
                <Link href="/dashboard" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Dashboard Pribadi
                </Link>
                <Link href="/dashboard/transactions" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname.startsWith('/dashboard/transactions') ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Pesanan Escrow
                </Link>
                <Link href="/dashboard/ledger" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/ledger' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  RekberPay Ledger
                </Link>
                <Link href="/dashboard/disputes" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/disputes' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  Sengketa
                </Link>
                <Link href="/dashboard/kyc" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/dashboard/kyc' ? 'text-blue-600 bg-blue-50/50 dark:bg-blue-950/30' : ''}`}>
                  KYC
                </Link>
              </>
            )
          ) : (
            <Link href="/" className={`px-3 py-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors ${pathname === '/' ? 'text-blue-600' : ''}`}>
              Landing Page
            </Link>
          )}
        </nav>

        {/* Action Controls */}
        <div className="flex items-center space-x-3">
          {mounted && isAuthenticated ? (
            <>
              {/* Wallet summary pill (Shopee-like quick balance display) */}
              <Link href="/dashboard/ledger" className="hidden sm:flex items-center space-x-2 px-3 py-1.5 rounded-xl border border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-800 hover:bg-zinc-100 transition-all">
                <Wallet className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                <div className="text-left">
                  <p className="text-[10px] text-zinc-400 leading-none">RekberPay</p>
                  <p className="text-xs font-extrabold text-zinc-900 dark:text-zinc-100">{formattedBalance}</p>
                </div>
              </Link>

              {/* Notification dropdown trigger */}
              <DropdownMenu
                trigger={
                  <Button variant="ghost" size="icon" className="relative rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800">
                    <Bell className="h-5 w-5 text-zinc-600 dark:text-zinc-400" />
                    {notifications.length > 0 && (
                      <span className="absolute top-1.5 right-1.5 h-2.5 w-2.5 rounded-full bg-blue-600 ring-2 ring-white dark:ring-zinc-900 animate-pulse" />
                    )}
                  </Button>
                }
              >
                <div className="p-3 w-80 text-sm">
                  <div className="flex items-center justify-between mb-3 pb-2 border-b border-zinc-100 dark:border-zinc-800">
                    <span className="font-bold text-xs uppercase tracking-wider text-zinc-700 dark:text-zinc-300">Notifikasi Pesanan & Escrow</span>
                    <Badge variant="secondary" className="text-[10px]">{notifications.length}</Badge>
                  </div>
                  <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                    {notifications.length === 0 ? (
                      <p className="text-xs text-zinc-400 text-center py-4">Belum ada notifikasi baru</p>
                    ) : (
                      notifications.map((n) => (
                        <div key={n.id} className="p-2.5 rounded-xl bg-zinc-50 dark:bg-zinc-800/60 hover:bg-zinc-100 transition-colors text-xs space-y-1">
                          <p className="font-semibold text-zinc-900 dark:text-zinc-100">{n.title}</p>
                          <p className="text-zinc-500 dark:text-zinc-400 leading-relaxed line-clamp-2">{n.message}</p>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </DropdownMenu>

              {/* User profile dropdown */}
              <DropdownMenu
                trigger={
                  <Button variant="ghost" className="flex items-center space-x-2 px-2 py-1 rounded-full hover:bg-zinc-100 dark:hover:bg-zinc-800">
                    <div className="h-8 w-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-bold text-xs shadow-xs">
                      {user?.fullName?.charAt(0) || user?.username?.charAt(0) || 'U'}
                    </div>
                    <span className="hidden md:inline-block text-xs font-semibold text-zinc-800 dark:text-zinc-200 max-w-[100px] truncate">
                      {user?.fullName || user?.username || 'Akun'}
                    </span>
                  </Button>
                }
              >
                <div className="px-3 py-2.5 border-b border-zinc-100 dark:border-zinc-800 space-y-0.5">
                  <p className="font-bold text-xs text-zinc-900 dark:text-zinc-100">
                    {user?.fullName || 'Pengguna'}
                  </p>
                  <p className="text-[10px] font-medium text-zinc-500 flex items-center gap-1">
                    <span className="px-1.5 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 uppercase">{user?.role || 'USER'}</span>
                    {user?.isKycVerified && <span className="text-emerald-600 font-bold">• KYC Verified</span>}
                  </p>
                </div>
                <DropdownMenuItem className="text-xs cursor-pointer" onClick={() => router.push('/dashboard/ledger')}>Dompet RekberPay</DropdownMenuItem>
                <DropdownMenuItem className="text-xs cursor-pointer" onClick={() => router.push('/dashboard/kyc')}>Verifikasi KYC</DropdownMenuItem>
                <DropdownMenuItem className="text-xs cursor-pointer" onClick={() => router.push('/dashboard/transactions')}>Daftar Pesanan</DropdownMenuItem>
                <DropdownMenuItem
                  onClick={handleLogout}
                  className="text-xs text-red-600 dark:text-red-400 flex items-center space-x-2 cursor-pointer font-semibold pt-2 border-t border-zinc-100 dark:border-zinc-800"
                >
                  <LogOut className="h-3.5 w-3.5" />
                  <span>Keluar Akun</span>
                </DropdownMenuItem>
              </DropdownMenu>
            </>
          ) : (
            <div className="flex items-center space-x-2">
              <Link href="/auth">
                <Button variant="default" size="sm" className="text-xs font-bold shadow-xs">
                  Masuk / Daftar
                </Button>
              </Link>
            </div>
          )}

          {/* Mobile hamburger menu */}
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden rounded-full"
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
        <div className="flex flex-col space-y-4 text-sm font-medium pt-4">
          {mounted && isAuthenticated ? (
            isAdmin ? (
              <>
                <Link
                  href="/dashboard/admin"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-purple-600 font-bold"
                >
                  Panel Admin
                </Link>
                <Link
                  href="/dashboard/disputes"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Moderasi Sengketa
                </Link>
                <Link
                  href="/dashboard/kyc"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Persetujuan KYC
                </Link>
                <Link
                  href="/dashboard/transactions"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Audit Pesanan
                </Link>
                <button
                  onClick={() => {
                    setMobileMenuOpen(false);
                    handleLogout();
                  }}
                  className="w-full text-left p-2 rounded-lg hover:bg-red-50 text-red-600 font-medium flex items-center space-x-2"
                >
                  <LogOut className="h-4 w-4" />
                  <span>Keluar</span>
                </button>
              </>
            ) : (
              <>
                <Link
                  href="/dashboard/catalog"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 text-blue-600 dark:text-blue-400 font-bold"
                >
                  Beranda (Katalog)
                </Link>
                <Link
                  href="/dashboard"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Dashboard Pribadi
                </Link>
                <Link
                  href="/dashboard/transactions"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Transaksi Escrow
                </Link>
                <Link
                  href="/dashboard/ledger"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Ledger Keuangan
                </Link>
                <Link
                  href="/dashboard/disputes"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Sengketa (Disputes)
                </Link>
                <Link
                  href="/dashboard/kyc"
                  onClick={() => setMobileMenuOpen(false)}
                  className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  Verifikasi KYC
                </Link>
                <button
                  onClick={() => {
                    setMobileMenuOpen(false);
                    handleLogout();
                  }}
                  className="w-full text-left p-2 rounded-lg hover:bg-red-50 text-red-600 font-medium flex items-center space-x-2"
                >
                  <LogOut className="h-4 w-4" />
                  <span>Keluar</span>
                </button>
              </>
            )
          ) : (
            <>
              <Link
                href="/"
                onClick={() => setMobileMenuOpen(false)}
                className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
              >
                Landing Page
              </Link>
              <div className="pt-2 border-t border-zinc-200 dark:border-zinc-800">
                <Link
                  href="/auth"
                  onClick={() => setMobileMenuOpen(false)}
                  className="block w-full text-center py-2.5 bg-blue-600 text-white rounded-xl text-xs font-bold"
                >
                  Masuk / Daftar Akun
                </Link>
              </div>
            </>
          )}
        </div>
      </Sheet>
    </header>
  );
}
