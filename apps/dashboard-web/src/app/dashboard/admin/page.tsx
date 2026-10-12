import React from 'react';
import { Navbar } from '@/components/modules/Navbar';
import { AdminPanelModule } from '@/components/modules/AdminPanelModule';

export default function AdminPageRoute() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8 space-y-8">
        <AdminPanelModule />
      </main>
    </div>
  );
}
