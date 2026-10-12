import 'server-only';
import { cookies, headers } from 'next/headers';
import { isJwtExpired } from './jwt';
import { JWT_COOKIE_NAME } from '@/lib/constants/auth';

/**
 * Retrieves the JWT token from secure cookies on the server side.
 */
export async function getServerJwtToken(): Promise<string | null> {
  try {
    const cookieStore = await cookies();
    let token = cookieStore.get(JWT_COOKIE_NAME)?.value;
    if (!token) {
      // Fallback check header Authorization
      const authHeader = (await headers()).get('authorization');
      if (authHeader && authHeader.startsWith('Bearer ')) {
        token = authHeader.substring(7);
      }
    }
    if (!token || isJwtExpired(token)) {
      return null;
    }
    return token;
  } catch {
    return null;
  }
}
