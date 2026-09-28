export const JWT_COOKIE_NAME =
  process.env.NODE_ENV === 'production'
    ? '__Host-rekberkuy-jwt'   // production: HTTPS, Secure=true
    : 'rekberkuy-jwt';          // development: HTTP localhost

export const CSRF_COOKIE_NAME =
  process.env.NODE_ENV === 'production'
    ? '__Host-rekberkuy-csrf'
    : 'rekberkuy-csrf';
