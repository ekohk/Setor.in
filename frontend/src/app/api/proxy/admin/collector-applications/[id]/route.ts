import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

async function tok(): Promise<string | undefined> {
  const s = await auth();
  return (s as Record<string, unknown> | null)?.accessToken as string | undefined;
}

// POST /api/proxy/admin/collector-applications/:id?action=approve|reject
export async function POST(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const token = await tok();
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const action = new URL(req.url).searchParams.get('action');
  if (!action) return NextResponse.json({ error: { code: 'BAD_REQUEST', message: 'action required' } }, { status: 400 });

  const body = await req.text();
  const upstream = await fetch(`${BASE}/v1/admin/collector-applications/${id}/${action}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: body || undefined,
    cache: 'no-store',
  });
  return NextResponse.json(await upstream.json(), { status: upstream.status });
}
