'use client';

import { logoutAction } from '@/actions/auth.actions';
import { useAuthStore } from '@/store/useAuthStore';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { LogOut } from 'lucide-react';

export function LogoutButton() {
  const logout = useAuthStore((state) => state.logout);
  const router = useRouter();

  const handleLogout = async () => {
    await logoutAction();
    logout();
    router.push('/auth');
    router.refresh();
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleLogout}
      className="flex items-center gap-2 text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
    >
      <LogOut className="w-4 h-4" />
      Keluar
    </Button>
  );
}
