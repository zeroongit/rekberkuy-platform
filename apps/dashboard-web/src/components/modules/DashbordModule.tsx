'use client';

import React from 'react';
import Link from 'next/link';
import {
  Users, ShoppingBag, Briefcase, Calendar, Star,
  AlertCircle, Clock, TrendingUp, Wallet, ArrowRight,
  Package, Shield, CheckCircle, XCircle,
  Store, UserCheck, ReceiptText,
} from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/useAuthStore';
import type { UserProfile } from '@/types';
import { formatRupiah } from '@/lib/utils/formatCurrency';

import { Card, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  sub?: string;
  accent?: 'blue' | 'green' | 'yellow' | 'red' | 'purple';
}

interface QuickActionProps {
  label: string;
  href: string;
  icon: React.ReactNode;
  variant?: 'default' | 'outline';
}

const accentMap = {
  blue:   'bg-blue-50 text-blue-600 dark:bg-blue-950 dark:text-blue-400',
  green:  'bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400',
  yellow: 'bg-amber-50 text-amber-600 dark:bg-amber-950 dark:text-amber-400',
  red:    'bg-red-50 text-red-600 dark:bg-red-950 dark:text-red-400',
  purple: 'bg-violet-50 text-violet-600 dark:bg-violet-950 dark:text-violet-400',
};

function StatCard({ title, value, icon, sub, accent = 'blue' }: StatCardProps) {
  return (
    <Card className="border border-zinc-200 dark:border-zinc-800 shadow-sm">
      <CardContent className="p-5 flex items-center gap-4">
        <div className={`p-3 rounded-xl ${accentMap[accent]}`}>
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-xs font-medium text-zinc-500 dark:text-zinc-400 truncate">{title}</p>
          <p className="text-xl font-bold text-zinc-900 dark:text-zinc-100 mt-0.5">{value}</p>
          {sub && <p className="text-xs text-zinc-400 dark:text-zinc-500 mt-0.5">{sub}</p>}
        </div>
      </CardContent>
    </Card>
  );
}

function QuickAction({ label, href, icon, variant = 'outline' }: QuickActionProps) {
  const router = useRouter();
  return (
    <Button
      variant={variant}
      className="w-full justify-between h-11 text-sm"
      onClick={() => router.push(href)}
    >
      <span className="flex items-center gap-2">
        {icon}
        {label}
      </span>
      <ArrowRight className="w-4 h-4 opacity-50" />
    </Button>
  );
}

function WalletCard({ balance, tier }: { balance: number; tier?: string }) {
  const router = useRouter();
  return (
    <Card className="bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-800 text-white border-0 shadow-lg relative overflow-hidden">
      <div className="absolute top-0 right-0 -mt-8 -mr-8 w-40 h-40 bg-white/10 rounded-full blur-2xl pointer-events-none" />
      <CardContent className="p-6 relative z-10">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-xl bg-white/20 backdrop-blur-md">
              <Wallet className="w-5 h-5 text-white" />
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-blue-100">Saldo Dompet RekberPay</p>
              <p className="text-[10px] text-blue-200">Aman & Terlindungi Escrow</p>
            </div>
          </div>
          {tier && (
            <Badge className="bg-white/20 text-white text-xs border-0 backdrop-blur-md font-bold px-3 py-1">
              <Star className="w-3.5 h-3.5 mr-1.5 fill-current" />
              {tier} VIP
            </Badge>
          )}
        </div>
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mt-2">
          <div>
            <p className="text-3xl font-black tracking-tight text-white">{formatRupiah(balance)}</p>
            <p className="text-xs text-blue-100 mt-1">Siap digunakan untuk transaksi barang, jasa, atau event</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              onClick={() => router.push('/dashboard/ledger')}
              className="bg-white text-blue-700 hover:bg-blue-50 font-bold text-xs h-9 px-4 shadow-sm"
            >
              Top-Up / Tarik
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function SectionHeader({ title, sub }: { title: string; sub?: string }) {
  return (
    <div className="mb-3">
      <h2 className="text-base font-semibold text-zinc-900 dark:text-zinc-100">{title}</h2>
      {sub && <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-0.5">{sub}</p>}
    </div>
  );
}

/** ADMIN */
function AdminDashboard({ user, stats }: { user: UserState; stats: { activeTransactions?: number; completedTransactions?: number } }) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
            Panel Admin
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-0.5">
            Selamat datang, {user.fullName || user.username}. Berikut ringkasan platform hari ini.
          </p>
        </div>
        <Badge className="bg-red-100 text-red-700 dark:bg-red-950 dark:text-red-400 border-0">
          <Shield className="w-3 h-3 mr-1" /> ADMIN
        </Badge>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard title="Total Pengguna" value="1.240" icon={<Users className="w-5 h-5" />} sub="↑ 12 minggu ini" accent="blue" />
        <StatCard title="KYC Menunggu" value="8" icon={<UserCheck className="w-5 h-5" />} sub="Perlu direview" accent="yellow" />
        <StatCard title="Sengketa Aktif" value="3" icon={<AlertCircle className="w-5 h-5" />} sub="Perlu mediasi" accent="red" />
        <StatCard title="Transaksi Aktif" value={stats.activeTransactions ?? 94} icon={<ReceiptText className="w-5 h-5" />} sub="Goods + Jasa + Event" accent="green" />
      </div>

      <div>
        <SectionHeader title="Aksi Cepat" />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <QuickAction label="Review KYC Pending" href="/dashboard/kyc" icon={<UserCheck className="w-4 h-4" />} variant="default" />
          <QuickAction label="Mediasi Sengketa" href="/dashboard/disputes" icon={<AlertCircle className="w-4 h-4" />} />
          <QuickAction label="Semua Transaksi" href="/dashboard/transactions" icon={<ReceiptText className="w-4 h-4" />} />
          <QuickAction label="Panel Admin" href="/dashboard/admin" icon={<Shield className="w-4 h-4" />} />
        </div>
      </div>
    </div>
  );
}

/** BUYER / USER */
function BuyerDashboard({ user, stats }: { user: UserState; stats: { walletBalance?: number; activeTransactions?: number; completedTransactions?: number } }) {
  const balance = stats.walletBalance !== undefined ? stats.walletBalance : (user.walletBalance || 0);
  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 bg-white dark:bg-zinc-900 p-6 rounded-2xl border border-zinc-200 dark:border-zinc-800 shadow-xs">
        <div>
          <h1 className="text-2xl font-black text-zinc-900 dark:text-zinc-100">
            Halo, {user.fullName || user.username} 👋
          </h1>
          <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-1">
            Semua transaksi barang, jasa, dan event dijamin aman dengan RekberPay Escrow.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs font-semibold px-3 py-1.5 bg-zinc-50 dark:bg-zinc-800">
            Status: {user.role}
          </Badge>
        </div>
      </div>

      <WalletCard balance={balance} tier={user.loyaltyTier} />

      {/* Shopee-like Order Status Quick Filters */}
      <Card className="p-5 border border-zinc-200 dark:border-zinc-800">
        <h3 className="text-sm font-bold text-zinc-900 dark:text-zinc-100 mb-4">Status Pesanan Saya</h3>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-center">
          <Link href="/dashboard/transactions" className="p-3 rounded-xl bg-zinc-50 dark:bg-zinc-800/60 hover:bg-blue-50 dark:hover:bg-blue-950/40 transition-colors group">
            <Clock className="w-6 h-6 mx-auto text-amber-500 group-hover:scale-110 transition-transform mb-1.5" />
            <p className="text-xs font-bold text-zinc-800 dark:text-zinc-200">Belum Bayar</p>
            <p className="text-[10px] text-zinc-400">Menunggu transfer</p>
          </Link>
          <Link href="/dashboard/transactions" className="p-3 rounded-xl bg-zinc-50 dark:bg-zinc-800/60 hover:bg-blue-50 dark:hover:bg-blue-950/40 transition-colors group">
            <Package className="w-6 h-6 mx-auto text-blue-600 group-hover:scale-110 transition-transform mb-1.5" />
            <p className="text-xs font-bold text-zinc-800 dark:text-zinc-200">Dikemas / Proses</p>
            <p className="text-[10px] text-zinc-400">Escrow terkunci</p>
          </Link>
          <Link href="/dashboard/transactions" className="p-3 rounded-xl bg-zinc-50 dark:bg-zinc-800/60 hover:bg-blue-50 dark:hover:bg-blue-950/40 transition-colors group">
            <CheckCircle className="w-6 h-6 mx-auto text-emerald-600 group-hover:scale-110 transition-transform mb-1.5" />
            <p className="text-xs font-bold text-zinc-800 dark:text-zinc-200">Selesai</p>
            <p className="text-[10px] text-zinc-400">Dana dirilis</p>
          </Link>
          <Link href="/dashboard/disputes" className="p-3 rounded-xl bg-zinc-50 dark:bg-zinc-800/60 hover:bg-blue-50 dark:hover:bg-blue-950/40 transition-colors group">
            <AlertCircle className="w-6 h-6 mx-auto text-red-500 group-hover:scale-110 transition-transform mb-1.5" />
            <p className="text-xs font-bold text-zinc-800 dark:text-zinc-200">Sengketa / Komplain</p>
            <p className="text-[10px] text-zinc-400">Mediasi admin</p>
          </Link>
        </div>
      </Card>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard title="Transaksi Aktif" value={stats.activeTransactions ?? 2} icon={<Clock className="w-5 h-5" />} sub="Dalam pengawasan" accent="yellow" />
        <StatCard title="Transaksi Selesai" value={stats.completedTransactions ?? 17} icon={<CheckCircle className="w-5 h-5" />} sub="Total berhasil" accent="green" />
        <StatCard title="Barang Dibeli" value="9" icon={<ShoppingBag className="w-5 h-5" />} sub="Fisik & Digital" accent="blue" />
        <StatCard title="Jasa & Event" value="8" icon={<Briefcase className="w-5 h-5" />} sub="Freelance & EO" accent="purple" />
      </div>

      {!user.isKycVerified && (
        <Card className="border border-amber-200 bg-amber-50/80 dark:border-amber-800 dark:bg-amber-950/30">
          <CardContent className="p-5 flex items-start gap-4">
            <AlertCircle className="w-6 h-6 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
            <div className="flex-1">
              <p className="text-sm font-bold text-amber-900 dark:text-amber-300">Akun Belum Terverifikasi KYC</p>
              <p className="text-xs text-amber-700 dark:text-amber-400 mt-1 leading-relaxed">
                Lengkapi verifikasi identitas (KTP & Selfie) untuk membuka limit transaksi hingga Rp 100.000.000 dan akses fitur merchant verified.
              </p>
              <Button size="sm" variant="default" onClick={() => window.location.href = '/dashboard/kyc'} className="mt-3 text-xs bg-amber-600 hover:bg-amber-700 text-white font-bold h-8">
                Mulai Verifikasi KYC Sekarang
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <div>
        <SectionHeader title="Akses Cepat & Layanan Escrow" sub="Pilih layanan transaksi aman RekberKuy" />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <QuickAction label="Buat Transaksi Escrow Baru" href="/dashboard/transactions" icon={<Package className="w-4 h-4 text-blue-600" />} variant="default" />
          <QuickAction label="Daftar Pesanan & Status" href="/dashboard/transactions" icon={<ReceiptText className="w-4 h-4 text-emerald-600" />} />
          <QuickAction label="Marketplace Vendor & EO" href="/dashboard/catalog" icon={<Calendar className="w-4 h-4 text-purple-600" />} />
          <QuickAction label="Ledger & Mutasi Keuangan" href="/dashboard/ledger" icon={<TrendingUp className="w-4 h-4 text-amber-600" />} />
        </div>
      </div>
    </div>
  );
}

/** SELLER / VENDOR / EVENT ORGANIZER */
function GeneralDashboard({ user, stats }: { user: UserState; stats: { walletBalance?: number; activeTransactions?: number; completedTransactions?: number } }) {
  const balance = stats.walletBalance !== undefined ? stats.walletBalance : (user.walletBalance || 0);
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
            Dashboard {user.role.replace('_', ' ')}
          </h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-400 mt-0.5">
            {user.fullName || user.username}
          </p>
        </div>
        {user.isKycVerified ? (
          <Badge className="bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-400 border-0">
            <CheckCircle className="w-3 h-3 mr-1" /> Terverifikasi
          </Badge>
        ) : (
          <Badge className="bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-400 border-0">
            <XCircle className="w-3 h-3 mr-1" /> Belum KYC
          </Badge>
        )}
      </div>

      <WalletCard balance={balance} tier={user.loyaltyTier} />

      <div className="grid grid-cols-2 gap-3">
        <StatCard title="Transaksi Aktif" value={stats.activeTransactions ?? 3} icon={<ShoppingBag className="w-5 h-5" />} accent="blue" />
        <StatCard title="Transaksi Selesai" value={stats.completedTransactions ?? 11} icon={<CheckCircle className="w-5 h-5" />} accent="green" />
        <StatCard title="Pending" value="2" icon={<Clock className="w-5 h-5" />} accent="yellow" />
        <StatCard title="Pendapatan" value={formatRupiah(8200000)} icon={<TrendingUp className="w-5 h-5" />} accent="purple" />
      </div>

      <div>
        <SectionHeader title="Aksi Cepat" sub="Kelola transaksi dan operasional kamu" />
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <QuickAction label="Daftar Transaksi" href="/dashboard/transactions" icon={<Package className="w-4 h-4" />} variant="default" />
          <QuickAction label="Marketplace" href="/dashboard/catalog" icon={<Store className="w-4 h-4" />} />
          <QuickAction label="Ledger Keuangan" href="/dashboard/ledger" icon={<TrendingUp className="w-4 h-4" />} />
          <QuickAction label="Sengketa" href="/dashboard/disputes" icon={<AlertCircle className="w-4 h-4" />} />
        </div>
      </div>
    </div>
  );
}

type UserState = UserProfile & {
  walletBalance?: number;
  isKycVerified?: boolean;
  loyaltyTier?: string;
};

export default function DashboardModule({
  initialUser,
  initialStats,
}: {
  initialUser?: UserProfile | null;
  initialStats?: { walletBalance?: number; activeTransactions?: number; completedTransactions?: number };
}) {
  const storeUser = useAuthStore((state) => state.user) as UserState | null;
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  const user: UserState | null = (initialUser as UserState) || storeUser;
  const stats = initialStats || {};
  const isAuth = isAuthenticated || !!initialUser;

  if (!isAuth || !user) {
    return (
      <div className="flex items-center justify-center min-h-[40vh]">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 rounded-full border-2 border-blue-600 border-t-transparent animate-spin" />
          <p className="text-sm text-zinc-500">Memuat dashboard...</p>
        </div>
      </div>
    );
  }

  const role = user.role;

  if (role === 'ADMIN') {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 py-6">
        <AdminDashboard user={user} stats={stats} />
      </div>
    );
  }

  if (role === 'USER' || role === 'VERIFIED_MERCHANT') {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 py-6">
        <BuyerDashboard user={user} stats={stats} />
      </div>
    );
  }

  return (
    <div className="w-full max-w-4xl mx-auto px-4 py-6">
      <GeneralDashboard user={user} stats={stats} />
    </div>
  );
}