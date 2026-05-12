/**
 * Catch-all route handler for next-auth.
 * Mounted at /api/auth/* (signin, callback, signout, session, csrf, ...).
 *
 * Auth.js `handlers` is `{ GET, POST }` — destructure & re-export here so
 * Next.js's App Router picks them up as route handlers.
 */
import { handlers } from '@/lib/auth';

export const { GET, POST } = handlers;
