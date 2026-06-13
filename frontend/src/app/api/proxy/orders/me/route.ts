import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@/lib/auth';

const BASE = process.env.BACKEND_URL ?? 'http://localhost:8000';

export async function GET(req: NextRequest) {
  const session = await auth();
  const token = (session as Record<string, unknown> | null)?.accessToken as string | undefined;
  if (!token) return NextResponse.json({ error: { code: 'UNAUTHORIZED', message: 'Not authenticated' } }, { status: 401 });

  const { searchParams } = new URL(req.url);
  const page = searchParams.get('page') ?? '1';
  const pageSize = searchParams.get('page_size') ?? '20';

  let upstream: Response;
  try {
    upstream = await fetch(`${BASE}/v1/orders/me?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    });
  } catch {
    return NextResponse.json(
      { error: { code: 'UPSTREAM_UNAVAILABLE', message: 'Backend API is not reachable' } },
      { status: 503 },
    );
  }

  const body = await upstream.json();
  return NextResponse.json(body, { status: upstream.status });
}
