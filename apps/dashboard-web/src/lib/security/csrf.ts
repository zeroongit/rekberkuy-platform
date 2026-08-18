import { cookies } from 'next/headers';

const CSRF_COOKIE_NAME = '__Host-rekberkuy-csrf';
export const CSRF_HEADER_NAME = 'x-csrf-token';

/**
 * Generates a cryptographically secure random token for CSRF protection using Web Crypto API.
 */
export function generateCsrfToken(): string {
  const array = new Uint8Array(32);
  globalThis.crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

/**
 * Sets the CSRF token in a secure cookie (Server Action / Server Component).
 */
export async function setCsrfCookie(): Promise<string> {
  const cookieStore = await cookies();
  let token = cookieStore.get(CSRF_COOKIE_NAME)?.value;
  
  if (!token) {
    token = generateCsrfToken();
    cookieStore.set({
      name: CSRF_COOKIE_NAME,
      value: token,
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      path: '/',
      maxAge: 60 * 60 * 24, // 24 hours
    });
  }
  
  return token;
}

/**
 * Verifies if the incoming request CSRF token matches the cookie token.
 */
export async function verifyCsrfToken(headerToken: string | null): Promise<boolean> {
  if (!headerToken) return false;
  const cookieStore = await cookies();
  const cookieToken = cookieStore.get(CSRF_COOKIE_NAME)?.value;
  
  if (!cookieToken) {
    return false;
  }
  
  return cookieToken === headerToken;
}
