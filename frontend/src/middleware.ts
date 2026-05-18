/**
 * Auth middleware — protects routes that require login.
 *
 * Public routes:  /, /login, /api/auth/*
 * Protected:      /home, /sell, /map, /tracking, /wallet, /profile, /collector/*, /admin/*
 *
 * Unauthenticated requests to a protected route get 302 → /login?callbackUrl=<original>.
 * Auth.js handles cookie verification.
 */

import { auth } from '@/lib/auth';
import { NextResponse } from 'next/server';

const PUBLIC_PREFIXES = ['/login', '/api/auth', '/api/proxy', '/_next', '/favicon.ico'];
const PUBLIC_EXACT = new Set(['/']);

export default auth((req) => {
  const { pathname } = req.nextUrl;

  if (PUBLIC_EXACT.has(pathname) || PUBLIC_PREFIXES.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  // Session expired (refresh token failed) — force re-login.
  if (req.auth?.error === 'RefreshAccessTokenError') {
    const url = new URL('/login', req.nextUrl.origin);
    url.searchParams.set('callbackUrl', pathname + req.nextUrl.search);
    return NextResponse.redirect(url);
  }

  if (!req.auth) {
    const url = new URL('/login', req.nextUrl.origin);
    url.searchParams.set('callbackUrl', pathname + req.nextUrl.search);
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
});

// Skip middleware on static assets so they don't pay the auth cost.
export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
