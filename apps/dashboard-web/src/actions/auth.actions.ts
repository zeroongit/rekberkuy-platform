'use server';

import { cookies } from 'next/headers';
import { createServerAction } from '@/lib/action-client';
import { loginSchema, registerSchema, LoginInput, RegisterInput } from '@/lib/validations/auth.schema';
import { JWT_COOKIE_NAME } from '@/lib/constants/auth';

/**
 * Secure Server Action for User Login.
 * Validates input with Zod, communicates with core-service backend, and sets HttpOnly Secure JWT cookie.
 */
export const loginAction = createServerAction(
  loginSchema,
  async (data: LoginInput) => {
    const apiBase = process.env.API_INTERNAL_URL || 'http://localhost:8080/api/v1';

    const res = await fetch(`${apiBase}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'Login gagal' }));
      throw new Error(errBody.error || 'Email atau password salah');
    }

    const result = (await res.json()) as { token?: string; access_token?: string; user: unknown };
    const token = result.token || result.access_token;

    if (!token) {
      throw new Error('Token tidak diterima dari server.');
    }
    
    // Set secure HttpOnly cookie
    const cookieStore = await cookies();
    cookieStore.set({
      name: JWT_COOKIE_NAME,
      value: token,
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: process.env.NODE_ENV === 'production' ? 'strict' : 'lax',
      path: '/',
      maxAge: 60 * 60 * 24, // 24 hours
    });

    return {
      user: result.user,
      token: token,
    };
  },
  { requireAuth: false, requireCsrf: false }
);

/**
 * Secure Server Action for User Registration.
 */
export const registerAction = createServerAction(
  registerSchema,
  async (data: RegisterInput) => {
    const apiBase = process.env.API_INTERNAL_URL || 'http://localhost:8080/api/v1';

    const res = await fetch(`${apiBase}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'Registrasi gagal' }));
      throw new Error(errBody.error || 'Gagal mendaftarkan akun');
    }

    return await res.json();
  },
  { requireAuth: false, requireCsrf: false }
);

/**
 * Server Action for User Logout.
 * Deletes the HttpOnly JWT session cookie.
 */
export async function logoutAction() {
  const cookieStore = await cookies();
  cookieStore.delete(JWT_COOKIE_NAME);
  return { success: true };
}
