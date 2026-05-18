// Client-side admin API — all calls via /api/proxy/admin/*

import type { AdminUser, CollectorApplication, PaginatedResponse, UserStatus, UserRole, ApplicationStatus } from '@/types/api';

async function proxyFetch<T>(url: string): Promise<T> {
  const res = await fetch(url, { cache: 'no-store' });
  const env = await res.json() as { data?: T; meta?: Record<string, unknown>; error?: { message: string } };
  if (!res.ok) throw new Error(env.error?.message ?? `Request failed (${res.status})`);
  return env.data as T;
}

async function proxyPatch<T>(url: string, body?: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'PATCH',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
    cache: 'no-store',
  });
  const env = await res.json() as { data?: T; error?: { message: string } };
  if (!res.ok) throw new Error(env.error?.message ?? `Request failed (${res.status})`);
  return env.data as T;
}

async function proxyPost<T>(url: string, body?: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
    cache: 'no-store',
  });
  const env = await res.json() as { data?: T; error?: { message: string } };
  if (!res.ok) throw new Error(env.error?.message ?? `Request failed (${res.status})`);
  return env.data as T;
}

async function listPaginated<T>(
  url: string,
  page = 1,
  pageSize = 20,
): Promise<PaginatedResponse<T>> {
  const res = await fetch(url, { cache: 'no-store' });
  const env = await res.json() as {
    data?: T[];
    meta?: { total?: number; page?: number; page_size?: number };
    error?: { message: string };
  };
  if (!res.ok) throw new Error(env.error?.message ?? 'Gagal memuat data');
  return {
    items: env.data ?? [],
    page: env.meta?.page ?? page,
    page_size: env.meta?.page_size ?? pageSize,
    total: env.meta?.total ?? 0,
  };
}

export interface UserListParams {
  page?: number;
  page_size?: number;
  q?: string;
  role?: UserRole;
  status?: UserStatus;
}

export function getAdminUsers(p: UserListParams = {}): Promise<PaginatedResponse<AdminUser>> {
  const qs = new URLSearchParams();
  if (p.page)      qs.set('page', String(p.page));
  if (p.page_size) qs.set('page_size', String(p.page_size));
  if (p.q)         qs.set('q', p.q);
  if (p.role)      qs.set('role', p.role);
  if (p.status)    qs.set('status', p.status);
  return listPaginated<AdminUser>(`/api/proxy/admin/users?${qs}`, p.page, p.page_size);
}

export const updateUserStatus = (id: string, status: UserStatus, reason?: string) =>
  proxyPatch<AdminUser>(`/api/proxy/admin/users/${id}?action=status`, { status, reason });

export const updateUserRole = (id: string, role: UserRole, reason?: string) =>
  proxyPatch<AdminUser>(`/api/proxy/admin/users/${id}?action=roles`, { role, reason });

export interface ApplicationListParams {
  page?: number;
  page_size?: number;
  status?: ApplicationStatus;
}

export function getCollectorApplications(p: ApplicationListParams = {}): Promise<PaginatedResponse<CollectorApplication>> {
  const qs = new URLSearchParams();
  if (p.page)      qs.set('page', String(p.page));
  if (p.page_size) qs.set('page_size', String(p.page_size));
  if (p.status)    qs.set('status', p.status);
  return listPaginated<CollectorApplication>(`/api/proxy/admin/collector-applications?${qs}`, p.page, p.page_size);
}

export const approveApplication = (id: string) =>
  proxyPost<CollectorApplication>(`/api/proxy/admin/collector-applications/${id}?action=approve`);

export const rejectApplication = (id: string, reason: string) =>
  proxyPost<CollectorApplication>(`/api/proxy/admin/collector-applications/${id}?action=reject`, { reason });

export { proxyFetch };
