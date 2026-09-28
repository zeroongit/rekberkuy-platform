import React from 'react';
import { Navbar } from '@/components/modules/Navbar';
import DashboardModule from '@/components/modules/DashbordModule';
import { getServerUser } from '@/lib/auth/server';
import { getDashboardStats } from '@/lib/api/dashboard';
import { redirect } from 'next/navigation';

export default async function DashboardOverviewRoute() {
  const user = await getServerUser();
  if (!user) {
    redirect('/auth');
  }
  if (user.role === 'ADMIN') {
    redirect('/dashboard/admin');
  }
  const stats = await getDashboardStats();

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        <DashboardModule initialUser={user} initialStats={stats} />
      </main>
    </div>
  );
}
