/**
 * Server-side fetch wrapper for calling the Setor.in backend.
 *
 * Use only in server components / route handlers / server actions —
 * it reads the session via next-auth's `auth()` and attaches the access_token
 * as Bearer. The token never leaves the server, so client-side JS can't read it.
 *
 * For client components, expose a thin endpoint at /api/proxy/* (TODO sprint
 * berikutnya) that forwards to backend with the same Bearer pattern.
 */

import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

export interface Envelope<T = unknown> {
  data?: T;
  meta?: Record<string, unknown>;
  error?: { code: string; message: string; details?: Record<string, unknown> };
}

interface FetchOpts extends Omit<RequestInit, 'body'> {
  /** Object — will be JSON.stringified and Content-Type set automatically. */
  json?: unknown;
  /** Skip auth (for /v1/materials etc.) */
  noAuth?: boolean;
}

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string, public details?: Record<string, unknown>) {
    super(message);
  }
}

export async function api<T = unknown>(path: string, opts: FetchOpts = {}): Promise<T> {
  const headers = new Headers(opts.headers);

  if (!opts.noAuth) {
    const session = await auth();
    const token = (session as Record<string, unknown> | null)?.accessToken as string | undefined;
    if (token) headers.set('Authorization', `Bearer ${token}`);
  }

  let body: BodyInit | undefined;
  if (opts.json !== undefined) {
    headers.set('Content-Type', 'application/json');
    body = JSON.stringify(opts.json);
  }

  const res = await fetch(BASE + path, {
    ...opts,
    headers,
    body,
    cache: 'no-store',
  });

  const envelope = (await res.json().catch(() => ({}))) as Envelope<T>;

  if (!res.ok) {
    const err = envelope.error;
    throw new ApiError(
      res.status,
      err?.code ?? 'UNKNOWN',
      err?.message ?? `Request failed (${res.status})`,
      err?.details,
    );
  }

  return envelope.data as T;
}
