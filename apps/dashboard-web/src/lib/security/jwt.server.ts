import 'server-only';
import { cookies } from 'next/headers';
import { isJwtExpired } from './jwt';

const JWT_COOKIE_NAME = '__Host-rekberkuy-jwt';

/**
 * Retrieves the JWT token from secure cookies on the server side.
 */
export async function getServerJwtToken(): Promise<string | null> {
  try {
    const cookieStore = await cookies();
    const token = cookieStore.get(JWT_COOKIE_NAME)?.value;
    if (!token || isJwtExpired(token)) {
      return null;
    }
    return token;
  } catch {
    return null;
  }
}
