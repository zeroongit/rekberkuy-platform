import React from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/Button';
import { Card, CardTitle, CardDescription } from '@/components/ui/Card';
import { ShieldCheck, ShoppingBag, Briefcase, Calendar, Lock, CheckCircle2 } from 'lucide-react';

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      {/* Header */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-200 dark:border-zinc-800 bg-white/80 dark:bg-zinc-900/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto flex h-16 items-center justify-between px-4 sm:px-6 lg:px-8">
          <div className="flex items-center space-x-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-600 text-white shadow-md">
              <ShieldCheck className="h-6 w-6" />
            </div>
            <span className="text-lg font-black tracking-tight text-zinc-900 dark:text-zinc-50">
              Rekber<span className="text-blue-600 dark:text-blue-400">Kuy</span>
            </span>
          </div>
          <div className="flex items-center space-x-3">
            <Link href="/auth">
              <Button variant="outline" size="sm" className="text-xs">
                Masuk / Daftar
              </Button>
            </Link>
            <Link href="/dashboard">
              <Button size="sm" className="text-xs">
                Buka Dashboard
              </Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-12 space-y-16">
        <div className="text-center max-w-3xl mx-auto space-y-4 pt-8">
          <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-blue-50 dark:bg-blue-950/50 text-blue-700 dark:text-blue-300 text-xs font-bold border border-blue-200 dark:border-blue-900">
            <Lock className="h-3.5 w-3.5" />
            <span>Platform Escrow Terdesentralisasi & Gasless Audit Log Avalanche</span>
          </div>
          <h1 className="text-4xl sm:text-6xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight leading-tight">
            Transaksi Aman untuk <span className="text-blue-600 dark:text-blue-400">Goods, Services, & Events</span>
          </h1>
          <p className="text-base sm:text-lg text-zinc-600 dark:text-zinc-400 leading-relaxed">
            RekberKuy adalah platform akun bersama (escrow) terpercaya di Indonesia. Setiap dana dikunci aman dalam sistem dan setiap rilis transaksi dicatat secara immutable sebagai audit log gasless di Avalanche C-Chain.
          </p>
          <div className="flex flex-wrap items-center justify-center gap-4 pt-4">
            <Link href="/dashboard">
              <Button size="lg" className="text-sm font-bold shadow-lg">
                Mulai Bertransaksi Sekarang
              </Button>
            </Link>
            <Link href="/auth">
              <Button variant="outline" size="lg" className="text-sm font-bold">
                Masuk Akun
              </Button>
            </Link>
          </div>
        </div>

        {/* Core Domains Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card className="p-6 space-y-3">
            <div className="w-12 h-12 rounded-2xl bg-blue-100 dark:bg-blue-950 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold">
              <ShoppingBag className="h-6 w-6" />
            </div>
            <CardTitle>Goods (Barang)</CardTitle>
            <CardDescription>
              Jual beli barang fisik & digital aman dengan kurir pilihan dan tenggat waktu konfirmasi otomatis.
            </CardDescription>
          </Card>

          <Card className="p-6 space-y-3">
            <div className="w-12 h-12 rounded-2xl bg-indigo-100 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-bold">
              <Briefcase className="h-6 w-6" />
            </div>
            <CardTitle>Services (Jasa)</CardTitle>
            <CardDescription>
              Proyek freelance & profesional berbasis pencairan bertahap (*milestones*) terverifikasi deliverable.
            </CardDescription>
          </Card>

          <Card className="p-6 space-y-3">
            <div className="w-12 h-12 rounded-2xl bg-purple-100 dark:bg-purple-950 text-purple-600 dark:text-purple-400 flex items-center justify-center font-bold">
              <Calendar className="h-6 w-6" />
            </div>
            <CardTitle>Events & Vendor Marketplace</CardTitle>
            <CardDescription>
              Procurement event organizer lengkap dengan alokasi vendor, invoice, dan pencairan multi-pihak.
            </CardDescription>
          </Card>
        </div>

        {/* Security & Architecture Highlights */}
        <div className="p-8 rounded-3xl bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 shadow-sm space-y-6">
          <h2 className="text-2xl font-black text-zinc-900 dark:text-zinc-100 text-center">
            Standar Keamanan 7-Layer Defense-in-Depth
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 text-sm">
            <div className="flex items-start space-x-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
              <div>
                <p className="font-bold">Zod Validation</p>
                <p className="text-xs text-zinc-500">Validasi input ganda klien & peladen.</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
              <div>
                <p className="font-bold">HttpOnly Secure Cookies</p>
                <p className="text-xs text-zinc-500">Mencegah XSS cookie theft & CSRF.</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
              <div>
                <p className="font-bold">Redis Cache Layer</p>
                <p className="text-xs text-zinc-500">Performa tinggi & idempotency.</p>
              </div>
            </div>
            <div className="flex items-start space-x-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0 mt-0.5" />
              <div>
                <p className="font-bold">Clean Architecture</p>
                <p className="text-xs text-zinc-500">Domain, UseCases, Repositories terisolasi.</p>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 py-6 text-center text-xs text-zinc-500">
        <p>© 2026 RekberKuy Platform. Built with Next.js v16, Tailwind CSS v4, and Clean Architecture.</p>
      </footer>
    </div>
  );
}
