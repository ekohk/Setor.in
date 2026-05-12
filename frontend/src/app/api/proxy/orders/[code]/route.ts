import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

async function bearerToken(): Promise<string | undefined> {
  const session = await auth();
  return (session as Record<string, unknown> | null)?.accessToken as string | undefined;
}

export async function GET(_req: NextRequest, { params }: { params: Promise<{ code: string }> }) {
  const { code } = await params;
  const token = await bearerToken();
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const upstream = await fetch(`${BASE}/v1/orders/${code}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  const body = await upstream.json();
  return NextResponse.json(body, { status: upstream.status });
}

export async function POST(req: NextRequest, { params }: { params: Promise<{ code: string }> }) {
  const { code } = await params;
  const token = await bearerToken();
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  // Routes: /api/proxy/orders/:code  (used for cancel and confirm-cash via searchParam action)
  const { searchParams } = new URL(req.url);
  const action = searchParams.get('action'); // 'cancel' | 'confirm-cash'
  const endpoint = action ? `${BASE}/v1/orders/${code}/${action}` : `${BASE}/v1/orders/${code}`;

  const body = await req.text();
  const upstream = await fetch(endpoint, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: body || undefined,
    cache: 'no-store',
  });
  const res = await upstream.json();
  return NextResponse.json(res, { status: upstream.status });
}
