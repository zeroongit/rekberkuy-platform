import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { decodeJwt, isJwtExpired } from '@/lib/security/jwt.utils';

const JWT_COOKIE_NAME = '__Host-rekberkuy-jwt';

export function middleware(request: NextRequest) {
  const token = request.cookies.get(JWT_COOKIE_NAME)?.value;
  const path = request.nextUrl.pathname;

  // 1. Check if token exists and is valid (not expired)
  if (!token || isJwtExpired(token)) {
    const url = new URL('/auth', request.url);
    url.searchParams.set('from', path);
    return NextResponse.redirect(url);
  }

  // 2. Role-Based Access Control (RBAC) for Admin routes
  if (path.startsWith('/admin') || path.startsWith('/dashboard/admin')) {
    const payload = decodeJwt(token);
    if (!payload || payload.role !== 'ADMIN') {
      const url = new URL('/dashboard', request.url);
      url.searchParams.set('error', 'unauthorized_admin');
      return NextResponse.redirect(url);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/dashboard/:path*',
    '/dashboard',
    '/transactions/:path*',
    '/transactions',
    '/ledger/:path*',
    '/ledger',
    '/sengketa/:path*',
    '/sengketa',
    '/marketplace/:path*',
    '/marketplace',
    '/kyc/:path*',
    '/kyc',
    '/admin/:path*',
    '/admin',
  ],
};
