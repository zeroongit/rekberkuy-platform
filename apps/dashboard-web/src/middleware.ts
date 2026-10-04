import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { jwtVerify } from 'jose';
import { JWT_COOKIE_NAME } from '@/lib/constants/auth';

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET || 'super-secret-jwt-key-change-in-production');

export async function middleware(request: NextRequest) {
  const token = request.cookies.get(JWT_COOKIE_NAME)?.value;
  const path = request.nextUrl.pathname;

  // 1. If visiting /auth while logged in, redirect based on role
  if (path.startsWith('/auth')) {
    if (token) {
      try {
        const { payload } = await jwtVerify(token, JWT_SECRET);
        const role = (payload as { role?: string })?.role;
        if (role === 'ADMIN') {
          return NextResponse.redirect(new URL('/dashboard/admin', request.url));
        }
        return NextResponse.redirect(new URL('/dashboard', request.url));
      } catch {
        return NextResponse.next();
      }
    }
    return NextResponse.next();
  }

  // 2. If visiting protected routes without token, redirect to /auth
  if (!token) {
    const url = new URL('/auth', request.url);
    url.searchParams.set('from', path);
    return NextResponse.redirect(url);
  }

  // 3. If token exists, verify role and enforce RBAC routing rules
  try {
    const { payload } = await jwtVerify(token, JWT_SECRET);
    const role = (payload as { role?: string })?.role;

    // Rule: Admin (ADMIN) must not access Landing Page (/) or public Beranda (/dashboard/catalog) or Buyer Dashboard (/dashboard).
    // Must be redirected exclusively to /dashboard/admin control center.
    if (role === 'ADMIN') {
      if (path === '/' || path === '/dashboard' || path === '/dashboard/catalog') {
        return NextResponse.redirect(new URL('/dashboard/admin', request.url));
      }
      return NextResponse.next();
    }

    // Rule: Non-admin users trying to access admin panel (/dashboard/admin or /admin)
    if (path.startsWith('/dashboard/admin') || path.startsWith('/admin')) {
      return NextResponse.redirect(new URL('/dashboard', request.url));
    }

    return NextResponse.next();
  } catch {
    const url = new URL('/auth', request.url);
    url.searchParams.set('from', path);
    const res = NextResponse.redirect(url);
    res.cookies.delete(JWT_COOKIE_NAME);
    return res;
  }
}

export const config = {
  matcher: [
    '/',
    '/auth',
    '/dashboard/:path*',
    '/dashboard',
    '/transactions/:path*',
    '/transactions',
    '/ledger/:path*',
    '/ledger',
    '/disputes/:path*',
    '/disputes',
    '/catalog/:path*',
    '/catalog',
    '/kyc/:path*',
    '/kyc',
    '/admin/:path*',
    '/admin',
  ],
};