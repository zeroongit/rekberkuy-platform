import { useAuthStore } from '@/store/useAuthStore';
import { JWTPayload, decodeJwt, isJwtExpired } from './jwt.utils';

export type { JWTPayload };
export { decodeJwt, isJwtExpired };

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
