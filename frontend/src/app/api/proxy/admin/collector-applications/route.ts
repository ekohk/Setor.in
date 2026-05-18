import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

async function tok(): Promise<string | undefined> {
  const s = await auth();
  return (s as Record<string, unknown> | null)?.accessToken as string | undefined;
}

// GET /api/proxy/admin/collector-applications?page=&page_size=&status=
export async function GET(req: NextRequest) {
  const token = await tok();
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const { searchParams } = new URL(req.url);
  const qs = new URLSearchParams();
  for (const key of ['page', 'page_size', 'status']) {
    const v = searchParams.get(key);
    if (v) qs.set(key, v);
  }

  const upstream = await fetch(`${BASE}/v1/admin/collector-applications?${qs}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  return NextResponse.json(await upstream.json(), { status: upstream.status });
}
