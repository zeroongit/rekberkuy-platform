/**
 * Sanitizes user input string to prevent XSS and HTML injection without relying on jsdom (safe for SSR).
 */
export function sanitizeString(input: string): string {
  if (typeof input !== 'string') return '';
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;');
}

/**
 * Sanitizes rich-text input where limited HTML formatting might be permitted.
 */
export function sanitizeRichText(input: string): string {
  if (typeof input !== 'string') return '';
  // Basic cleanup for allowed tags or fallback to sanitizeString
  return sanitizeString(input);
}

/**
 * Recursively sanitizes object string values.
 */
export function sanitizeObject<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    if (typeof obj === 'string') {
      return sanitizeString(obj) as unknown as T;
    }
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => sanitizeObject(item)) as unknown as T;
  }

  const sanitized = {} as Record<string, unknown>;
  for (const key of Object.keys(obj as object)) {
    sanitized[key] = sanitizeObject((obj as Record<string, unknown>)[key]);
  }

  return sanitized as T;
}
