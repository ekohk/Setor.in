import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

export async function GET(req: NextRequest) {
  const session = await auth();
  const tok = (session as Record<string, unknown> | null)?.accessToken as string | undefined;
  if (!tok) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const { searchParams } = new URL(req.url);
  const qs = new URLSearchParams({ page: searchParams.get('page') ?? '1', page_size: searchParams.get('page_size') ?? '20' });
  const upstream = await fetch(`${BASE}/v1/collector/orders/incoming?${qs}`, {
    headers: { Authorization: `Bearer ${tok}` },
    cache: 'no-store',
  });
  return NextResponse.json(await upstream.json(), { status: upstream.status });
}
