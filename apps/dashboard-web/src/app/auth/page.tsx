import React from 'react';
import { Navbar } from '@/components/modules/Navbar';
import { AuthModule } from '@/components/modules/AuthModule';

export default function AuthPageRoute() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950 flex flex-col">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-12 flex items-center justify-center">
        <AuthModule />
      </main>
    </div>
  );
}
