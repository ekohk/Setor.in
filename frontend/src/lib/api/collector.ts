// Client-side collector order API — all calls via /api/proxy/collector/*

import type { Order, PaginatedResponse } from '@/types/api';

async function proxyPost<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
    cache: 'no-store',
  });
  const env = await res.json() as { data?: T; error?: { message: string } };
  if (!res.ok) throw new Error(env.error?.message ?? `Request failed (${res.status})`);
  return env.data as T;
}

async function listOrders(url: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Order>> {
  const res = await fetch(`${url}?page=${page}&page_size=${pageSize}`, { cache: 'no-store' });
  const env = await res.json() as {
    data?: Order[];
    meta?: { total?: number; page?: number; page_size?: number };
    error?: { message: string };
  };
  if (!res.ok) throw new Error(env.error?.message ?? 'Gagal memuat orders');
  return {
    items: env.data ?? [],
    page: env.meta?.page ?? page,
    page_size: env.meta?.page_size ?? pageSize,
    total: env.meta?.total ?? 0,
  };
}

export const getIncomingOrders = (page?: number, pageSize?: number) =>
  listOrders('/api/proxy/collector/orders/incoming', page, pageSize);

export const getMyCollectorOrders = (page?: number, pageSize?: number) =>
  listOrders('/api/proxy/collector/orders/me', page, pageSize);

export const acceptOrder = (code: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=accept`);

export const startPickup = (code: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=start-pickup`);

export const arriveOrder = (code: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=arrive`);

export const verifyOTP = (code: string, otp_code: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=verify-otp`, { otp_code });

export const weighOrder = (code: string, actual_weight_kg: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=weigh`, { actual_weight_kg });

export const qualityOrder = (code: string, grade: 'A' | 'B' | 'C', notes?: string) =>
  proxyPost<Order>(`/api/proxy/collector/orders/${code}?action=quality`, { grade, notes });
