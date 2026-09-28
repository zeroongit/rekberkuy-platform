'use client';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { loginSchema, registerSchema, LoginInput, RegisterInput } from '@/lib/validations/auth.schema';
import { loginAction, registerAction } from '@/actions/auth.actions';
import { useServerAction } from '@/hooks/useServerAction';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { FormErrorAlert } from '@/components/ui/FormErrorAlert';
import { FieldError } from '@/components/ui/FieldError';
import { notify } from '@/components/providers/ToastProvider';
import { useAuthStore } from '@/store/useAuthStore';
import { UserProfile } from '@/types';

import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';

export function AuthModule() {
  const [isLogin, setIsLogin] = useState(true);
  const { setAuth } = useAuthStore();
  const router = useRouter();

  const loginForm = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const registerForm = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: '', password: '', fullName: '', phone: '', accountType: 'personal', role: 'USER' },
  });

  const accountType = registerForm.watch('accountType');

  const loginActionHook = useServerAction(loginSchema, loginAction);
  const registerActionHook = useServerAction(registerSchema, registerAction);

  const onLoginSubmit = async (values: LoginInput) => {
    const res = await loginActionHook.execute(values);
    if (res.success && res.data) {
      const data = res.data as { user: UserProfile; token: string };
      setAuth(data.user, data.token);
      notify.success('Login berhasil! Selamat datang.');

      const params = new URLSearchParams(window.location.search);
      const rawFrom = params.get('from');
      const role = data.user.role;
      let redirectTo = '/dashboard';
      if (role === 'ADMIN') {
        redirectTo = '/dashboard/admin';
      } else if (['EVENT_ORGANIZER', 'VERIFIED_MERCHANT', 'VERIFIED_VENDOR', 'SELLER', 'SERVICE_PROVIDER'].includes(role)) {
        redirectTo = '/dashboard/kyc'; // Arahkan ke modul pengisian data bisnis / KYC untuk akun komersial
      } else if (rawFrom && rawFrom.startsWith('/') && !rawFrom.startsWith('//') && rawFrom !== '/auth' && !rawFrom.startsWith('/dashboard/admin') && !rawFrom.startsWith('/admin')) {
        redirectTo = rawFrom;
      }
      window.location.assign(redirectTo);
    }
  };

  const onRegisterSubmit = async (values: RegisterInput) => {
    const payload = {
      ...values,
      role: values.accountType === 'commercial' ? (values.role || 'EVENT_ORGANIZER') : 'USER',
    };
    const res = await registerActionHook.execute(payload);
    if (res.success) {
      if (values.accountType === 'commercial') {
        notify.success('Registrasi akun komersial berhasil! Silakan login untuk melanjutkan ke verifikasi KYC & data bisnis.');
      } else {
        notify.success('Registrasi akun personal berhasil! Silakan login.');
      }
      setIsLogin(true);
      registerForm.reset();
    }
  };

  return (
    <div className="space-y-4 max-w-md mx-auto">
      <div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.push('/')}
          className="text-xs text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-zinc-100 p-0 h-auto font-semibold flex items-center gap-1.5"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Kembali ke Beranda Utama</span>
        </Button>
      </div>
      <Card className="p-6">
        <CardHeader className="px-0 pt-0">
          <CardTitle className="text-xl">{isLogin ? 'Masuk ke RekberKuy' : 'Daftar Akun Baru'}</CardTitle>
          <CardDescription className="text-xs mt-1">
            Dilindungi oleh Zod validation, CSRF verification, dan HttpOnly Secure cookies
          </CardDescription>
        </CardHeader>
        <CardContent className="px-0 pb-0 space-y-4">
        {isLogin ? (
          <form onSubmit={loginForm.handleSubmit(onLoginSubmit)} className="space-y-4">
            <FormErrorAlert message={loginActionHook.error} errors={loginActionHook.validationErrors} />

            <div>
              <label className="block text-xs font-bold mb-1">Email</label>
              <Input type="email" {...loginForm.register('email')} placeholder="user@rekberkuy.test" />
              <FieldError message={loginForm.formState.errors.email?.message} />
            </div>

            <div>
              <label className="block text-xs font-bold mb-1">Password</label>
              <Input type="password" {...loginForm.register('password')} placeholder="••••••••" />
              <FieldError message={loginForm.formState.errors.password?.message} />
            </div>

            <Button type="submit" disabled={loginActionHook.isPending} className="w-full text-xs font-bold">
              {loginActionHook.isPending ? 'Memproses...' : 'Masuk'}
            </Button>
          </form>
        ) : (
          <form onSubmit={registerForm.handleSubmit(onRegisterSubmit)} className="space-y-4">
            <FormErrorAlert message={registerActionHook.error} errors={registerActionHook.validationErrors} />

            <div>
              <label className="block text-xs font-bold mb-1">Tipe Akun</label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => registerForm.setValue('accountType', 'personal')}
                  className={`py-2 px-3 text-xs font-bold rounded-lg border transition-all ${
                    accountType === 'personal'
                      ? 'border-blue-600 bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
                      : 'border-zinc-200 text-zinc-600 dark:border-zinc-800 dark:text-zinc-400'
                  }`}
                >
                  Personal (Ritel)
                </button>
                <button
                  type="button"
                  onClick={() => registerForm.setValue('accountType', 'commercial')}
                  className={`py-2 px-3 text-xs font-bold rounded-lg border transition-all ${
                    accountType === 'commercial'
                      ? 'border-blue-600 bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
                      : 'border-zinc-200 text-zinc-600 dark:border-zinc-800 dark:text-zinc-400'
                  }`}
                >
                  Komersial (Bisnis/EO)
                </button>
              </div>
            </div>

            {accountType === 'commercial' && (
              <div>
                <label className="block text-xs font-bold mb-1">Peran Bisnis / Komersial</label>
                <select
                  {...registerForm.register('role')}
                  className="w-full text-xs p-2 rounded-md border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900"
                >
                  <option value="EVENT_ORGANIZER">Event Organizer (EO)</option>
                  <option value="VERIFIED_MERCHANT">Merchant / Seller Barang</option>
                  <option value="VERIFIED_VENDOR">Vendor Pendukung Event</option>
                  <option value="SERVICE_PROVIDER">Penyedia Jasa (Freelancer)</option>
                </select>
              </div>
            )}

            <div>
              <label className="block text-xs font-bold mb-1">Nama Lengkap</label>
              <Input type="text" {...registerForm.register('fullName')} placeholder="Budi Santoso" />
              <FieldError message={registerForm.formState.errors.fullName?.message} />
            </div>

            <div>
              <label className="block text-xs font-bold mb-1">Email</label>
              <Input type="email" {...registerForm.register('email')} placeholder="user@rekberkuy.test" />
              <FieldError message={registerForm.formState.errors.email?.message} />
            </div>

            <div>
              <label className="block text-xs font-bold mb-1">Password</label>
              <Input type="password" {...registerForm.register('password')} placeholder="••••••••" />
              <FieldError message={registerForm.formState.errors.password?.message} />
            </div>

            <Button type="submit" disabled={registerActionHook.isPending} className="w-full text-xs font-bold">
              {registerActionHook.isPending ? 'Memproses...' : 'Daftar Akun'}
            </Button>
          </form>
        )}

        <div className="flex items-center justify-center pt-2 border-t border-zinc-100 dark:border-zinc-800 text-xs">
          <button
            type="button"
            onClick={() => setIsLogin(!isLogin)}
            className="text-blue-600 dark:text-blue-400 font-medium hover:underline"
          >
            {isLogin ? 'Belum punya akun? Daftar' : 'Sudah punya akun? Masuk'}
          </button>
        </div>
      </CardContent>
    </Card>
    </div>
  );
}
