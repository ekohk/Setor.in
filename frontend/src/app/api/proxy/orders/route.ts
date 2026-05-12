import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

export async function POST(req: NextRequest) {
  const session = await auth();
  const token = (session as Record<string, unknown> | null)?.accessToken as string | undefined;
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const body = await req.text();
  const upstream = await fetch(`${BASE}/v1/orders`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body,
    cache: 'no-store',
  });
  const res = await upstream.json();
  return NextResponse.json(res, { status: upstream.status });
}
