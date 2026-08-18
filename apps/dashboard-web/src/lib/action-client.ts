import { z } from 'zod';
import { verifyCsrfToken } from '@/lib/security/csrf';
import { getServerJwtToken } from '@/lib/security/jwt.server';
import { sanitizeObject } from '@/lib/security/sanitize';

export interface ServerActionContext {
  token: string | null;
  authHeaders: Record<string, string>;
}

/**
 * Higher-order wrapper for Next.js Server Actions with built-in:
 * - CSRF verification
 * - JWT auth extraction
 * - Zod validation
 * - Secure error sanitization
 */
export function createServerAction<TSchema extends z.ZodTypeAny, TResult>(
  schema: TSchema,
  handler: (data: z.infer<TSchema>, ctx: ServerActionContext) => Promise<TResult>,
  options: { requireAuth?: boolean; requireCsrf?: boolean } = { requireAuth: true, requireCsrf: true }
) {
  return async (rawInput: z.infer<TSchema>, csrfTokenHeader?: string | null) => {
    try {
      // 1. CSRF Verification
      if (options.requireCsrf !== false) {
        const isValidCsrf = await verifyCsrfToken(csrfTokenHeader || null);
        if (!isValidCsrf) {
          return {
            success: false,
            error: 'CSRF token tidak valid atau kadaluarsa.',
          };
        }
      }

      // 2. JWT & Auth headers
      const token = await getServerJwtToken();
      const authHeaders: Record<string, string> = token ? { Authorization: `Bearer ${token}` } : {};

      if (options.requireAuth !== false && !token) {
        return {
          success: false,
          error: 'Autentikasi diperlukan. Silakan login kembali.',
        };
      }

      // 3. Sanitization & Zod validation
      const sanitized = sanitizeObject(rawInput);
      const parsed = schema.safeParse(sanitized);

      if (!parsed.success) {
        return {
          success: false,
          error: 'Validasi data gagal.',
          validationErrors: parsed.error.flatten().fieldErrors,
        };
      }

      // 4. Execute handler
      const data = await handler(parsed.data, { token, authHeaders });

      return {
        success: true,
        data,
      };
    } catch (err: unknown) {
      // Safe error logging & message sanitization (do not leak stack trace)
      const errMessage = err instanceof Error ? err.message : 'Terjadi kesalahan internal server.';
      console.error('[ServerAction Error]:', err);

      return {
        success: false,
        error: errMessage,
      };
    }
  };
}
