import DOMPurify from 'isomorphic-dompurify';

/**
 * Sanitizes user input string using isomorphic-dompurify to prevent XSS and HTML injection.
 */
export function sanitizeString(input: string): string {
  if (typeof input !== 'string') return '';
  return DOMPurify.sanitize(input, { ALLOWED_TAGS: [], ALLOWED_ATTR: [] });
}

/**
 * Sanitizes rich-text input where limited HTML formatting might be permitted.
 */
export function sanitizeRichText(input: string): string {
  if (typeof input !== 'string') return '';
  return DOMPurify.sanitize(input, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br', 'ul', 'ol', 'li'],
    ALLOWED_ATTR: [],
  });
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
