import { useAuthStore } from '@/store/useAuthStore';

export interface JWTPayload {
  sub: string;
  email: string;
  role: string;
  exp: number;
  iat: number;
}

/**
 * Parses and decodes a JWT token payload without external heavy libraries (base64url decode).
 */
export function decodeJwt(token: string): JWTPayload | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    
    const payload = parts[1];
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    
    return JSON.parse(jsonPayload) as JWTPayload;
  } catch {
    return null;
  }
}

/**
 * Checks if a JWT token is expired.
 */
export function isJwtExpired(token: string): boolean {
  const payload = decodeJwt(token);
  if (!payload || !payload.exp) return true;
  
  const currentTime = Math.floor(Date.now() / 1000);
  return payload.exp < currentTime;
}

/**
 * Creates Authorization header object with Bearer token for client-side API calls.
 */
export async function getAuthHeaders(): Promise<Record<string, string>> {
  let token: string | null = null;

  if (typeof window !== 'undefined') {
    token = useAuthStore.getState().token;
    if (token && isJwtExpired(token)) {
      token = null;
    }
  }

  if (!token) return {};
  
  return {
    'Authorization': `Bearer ${token}`,
  };
}
