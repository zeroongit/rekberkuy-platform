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

export function AuthModule() {
  const [isLogin, setIsLogin] = useState(true);
  const { setAuth } = useAuthStore();

  const loginForm = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const registerForm = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: '', password: '', fullName: '', phone: '' },
  });

  const loginActionHook = useServerAction(loginSchema, loginAction);
  const registerActionHook = useServerAction(registerSchema, registerAction);

  const onLoginSubmit = async (values: LoginInput) => {
    const res = await loginActionHook.execute(values);
    if (res.success && res.data) {
      const data = res.data as { user: UserProfile; token: string };
      setAuth(data.user, data.token);
      notify.success('Login berhasil! Selamat datang.');
    }
  };

  const onRegisterSubmit = async (values: RegisterInput) => {
    const res = await registerActionHook.execute(values);
    if (res.success) {
      notify.success('Registrasi berhasil! Silakan login.');
      setIsLogin(true);
      registerForm.reset();
    }
  };

  return (
    <Card className="max-w-md mx-auto p-6">
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
  );
}
