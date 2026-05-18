import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

async function token(): Promise<string | undefined> {
  const session = await auth();
  return (session as Record<string, unknown> | null)?.accessToken as string | undefined;
}

export async function GET() {
  const tok = await token();
  if (!tok) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const upstream = await fetch(`${BASE}/v1/users/me`, {
    headers: { Authorization: `Bearer ${tok}` },
    cache: 'no-store',
  });
  const body = await upstream.json();
  return NextResponse.json(body, { status: upstream.status });
}

export async function PATCH(req: NextRequest) {
  const tok = await token();
  if (!tok) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const body = await req.text();
  const upstream = await fetch(`${BASE}/v1/users/me`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${tok}`, 'Content-Type': 'application/json' },
    body,
    cache: 'no-store',
  });
  const res = await upstream.json();
  return NextResponse.json(res, { status: upstream.status });
}
