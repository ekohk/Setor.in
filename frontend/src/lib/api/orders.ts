// Client-side order API — all calls go through /api/proxy/* so the
// Bearer token never touches client JS.

import type { Order, PaginatedResponse } from '@/types/api';

async function proxyFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, { ...init, cache: 'no-store' });
  const envelope = await res.json() as { data?: T; meta?: Record<string, unknown>; error?: { message: string } };
  if (!res.ok) throw new Error(envelope.error?.message ?? `Request failed (${res.status})`);
  return envelope.data as T;
}

export interface CreateOrderPayload {
  material_slug: string;
  estimated_weight_kg: string;
  method: 'pickup' | 'dropoff';
  address_text: string;
  latitude?: number;
  longitude?: number;
  notes?: string;
}

export async function createOrder(payload: CreateOrderPayload): Promise<Order> {
  const res = await fetch('/api/proxy/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    cache: 'no-store',
  });
  const envelope = await res.json() as { data?: Order; error?: { message: string } };
  if (!res.ok) throw new Error(envelope.error?.message ?? 'Gagal membuat order');
  return envelope.data as Order;
}

export async function getMyOrders(page = 1, pageSize = 20): Promise<PaginatedResponse<Order>> {
  const res = await fetch(`/api/proxy/orders/me?page=${page}&page_size=${pageSize}`, { cache: 'no-store' });
  const envelope = await res.json() as {
    data?: Order[];
    meta?: { total?: number; page?: number; page_size?: number; total_pages?: number };
    error?: { message: string };
  };
  if (!res.ok) throw new Error(envelope.error?.message ?? 'Gagal memuat orders');
  const meta = envelope.meta ?? {};
  return {
    items: envelope.data ?? [],
    page: meta.page ?? page,
    page_size: meta.page_size ?? pageSize,
    total: meta.total ?? 0,
  };
}

export async function getOrderByCode(code: string): Promise<Order> {
  return proxyFetch<Order>(`/api/proxy/orders/${code}`);
}

export async function cancelOrder(code: string, reason: string): Promise<Order> {
  return proxyFetch<Order>(`/api/proxy/orders/${code}?action=cancel`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ reason }),
  });
}

export async function confirmCash(code: string): Promise<Order> {
  return proxyFetch<Order>(`/api/proxy/orders/${code}?action=confirm-cash`, { method: 'POST' });
}
