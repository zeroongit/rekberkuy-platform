import 'server-only';
import { cookies,} from 'next/headers';
import { jwtVerify } from 'jose';
import { JWT_COOKIE_NAME } from '@/lib/constants/auth';
import { UserProfile } from '@/types';

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET);

export async function getServerUser(): Promise<UserProfile | null> {
  try {
    const cookieStore = await cookies();
    const token = cookieStore.get(JWT_COOKIE_NAME)?.value;
    if (!token) return null;

    const { payload } = await jwtVerify(token, JWT_SECRET);
    return payload as unknown as UserProfile;
  } catch {
    // token invalid atau expired
    return null;
  }
}
