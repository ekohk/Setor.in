'use client';

import { useState, useEffect, useCallback } from 'react';
import type { AdminUser, UserStatus, UserRole } from '@/types/api';
import { USER_STATUS_LABEL, USER_ROLE_LABEL } from '@/types/api';
import { getAdminUsers, updateUserStatus, updateUserRole } from '@/lib/api/admin';
import Link from 'next/link';

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(iso));
}

const STATUS_COLOR: Record<UserStatus, string> = {
  pending_verification: 'bg-amber-50 text-amber-700',
  active:               'bg-accent-soft text-accent-deep',
  suspended:            'bg-red-50 text-red-600',
  deleted:              'bg-paper-2 text-ink-4',
};

const ROLE_COLOR: Record<UserRole, string> = {
  user:        'bg-paper-2 text-ink-3',
  cv:          'bg-emerald-50 text-emerald-700',
  collector:   'bg-blue-50 text-blue-700',
  admin:       'bg-purple-50 text-purple-700',
  super_admin: 'bg-amber-50 text-amber-700',
};

type Modal =
  | { type: 'status'; user: AdminUser }
  | { type: 'role';   user: AdminUser }
  | null;

export default function AdminUserList() {
  const [users, setUsers]     = useState<AdminUser[]>([]);
  const [total, setTotal]     = useState(0);
  const [page, setPage]       = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState('');

  const [q, setQ]           = useState('');
  const [roleF, setRoleF]   = useState<UserRole | ''>('');
  const [statusF, setStatusF] = useState<UserStatus | ''>('');

  const [modal, setModal]   = useState<Modal>(null);
  const [busy, setBusy]     = useState(false);
  const [actionErr, setActionErr] = useState('');

  // Status modal state
  const [newStatus, setNewStatus]       = useState<UserStatus>('active');
  const [statusReason, setStatusReason] = useState('');
  // Role modal state
  const [newRole, setNewRole]     = useState<UserRole>('user');
  const [roleReason, setRoleReason] = useState('');

  const PAGE_SIZE = 20;

  const load = useCallback(async (p: number) => {
    setLoading(true);
    setError('');
    try {
      const res = await getAdminUsers({
        page: p, page_size: PAGE_SIZE,
        q: q || undefined,
        role: roleF || undefined,
        status: statusF || undefined,
      });
      setUsers(res.items);
      setTotal(res.total);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal memuat data');
    } finally {
      setLoading(false);
    }
  }, [q, roleF, statusF]);

  useEffect(() => {
    setPage(1);
    load(1);
  }, [load]);

  function openStatusModal(u: AdminUser) {
    setNewStatus(u.status === 'active' ? 'suspended' : 'active');
    setStatusReason('');
    setActionErr('');
    setModal({ type: 'status', user: u });
  }

  function openRoleModal(u: AdminUser) {
    setNewRole(u.primary_role);
    setRoleReason('');
    setActionErr('');
    setModal({ type: 'role', user: u });
  }

  async function submitStatus() {
    if (!modal || modal.type !== 'status') return;
    setBusy(true);
    setActionErr('');
    try {
      const updated = await updateUserStatus(modal.user.id, newStatus, statusReason || undefined);
      setUsers(prev => prev.map(u => u.id === updated.id ? updated : u));
      setModal(null);
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal mengubah status');
    } finally {
      setBusy(false);
    }
  }

  async function submitRole() {
    if (!modal || modal.type !== 'role') return;
    setBusy(true);
    setActionErr('');
    try {
      const updated = await updateUserRole(modal.user.id, newRole, roleReason || undefined);
      setUsers(prev => prev.map(u => u.id === updated.id ? updated : u));
      setModal(null);
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal mengubah role');
    } finally {
      setBusy(false);
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-5">
        <h1 className="text-white text-2xl font-bold tracking-tight">Manajemen User</h1>
        <p className="text-white/70 text-sm mt-1">Total {total} pengguna terdaftar</p>
      </div>

      {/* Filters */}
      <div className="bg-surface border-b border-line px-4 py-3 space-y-2 sticky top-0 z-10">
        <input
          type="search"
          placeholder="Cari nama / email..."
          value={q}
          onChange={e => setQ(e.target.value)}
          className="w-full border border-line rounded-xl px-4 py-2.5 text-[13px] focus:outline-none focus:border-accent"
        />
        <div className="flex gap-2">
          <select
            value={roleF}
            onChange={e => setRoleF(e.target.value as UserRole | '')}
            className="flex-1 border border-line rounded-xl px-3 py-2 text-[12px] focus:outline-none focus:border-accent bg-surface"
          >
            <option value="">Semua Role</option>
            <option value="user">Pengguna</option>
            <option value="cv">CV Partner</option>
            <option value="collector">Collector</option>
            <option value="admin">Admin</option>
            <option value="super_admin">Super Admin</option>
          </select>
          <select
            value={statusF}
            onChange={e => setStatusF(e.target.value as UserStatus | '')}
            className="flex-1 border border-line rounded-xl px-3 py-2 text-[12px] focus:outline-none focus:border-accent bg-surface"
          >
            <option value="">Semua Status</option>
            <option value="pending_verification">Belum Verifikasi</option>
            <option value="active">Aktif</option>
            <option value="suspended">Disuspend</option>
          </select>
        </div>
      </div>

      <div className="px-4 py-4">
        {loading && (
          <div className="flex flex-col items-center py-16 text-ink-4">
            <div className="w-8 h-8 border-2 border-accent/30 border-t-accent rounded-full animate-spin" />
            <p className="mt-3 text-[13px]">Memuat data...</p>
          </div>
        )}

        {!loading && error && (
          <div className="glass flex flex-col items-center py-10 text-center">
            <p className="text-[13px] text-red-600">{error}</p>
            <button onClick={() => load(page)} className="mt-3 text-[12px] font-semibold text-accent">Coba Lagi</button>
          </div>
        )}

        {!loading && !error && users.length === 0 && (
          <div className="glass flex flex-col items-center py-12 text-center mt-2">
            <p className="text-[13px] font-semibold text-ink-2">Tidak ada pengguna ditemukan</p>
          </div>
        )}

        {!loading && !error && users.length > 0 && (
          <div className="space-y-2">
            {users.map(u => (
              <div key={u.id} className="glass px-4 py-3">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-[14px] font-bold text-ink truncate">{u.full_name}</p>
                    <p className="text-[12px] text-ink-3 truncate">{u.email}</p>
                    <p className="text-[11px] text-ink-4 mt-0.5">{formatDate(u.created_at)}</p>
                  </div>
                  <div className="flex flex-col items-end gap-1 shrink-0">
                    <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${STATUS_COLOR[u.status]}`}>
                      {USER_STATUS_LABEL[u.status]}
                    </span>
                    <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${ROLE_COLOR[u.primary_role]}`}>
                      {USER_ROLE_LABEL[u.primary_role]}
                    </span>
                  </div>
                </div>
                <div className="flex gap-2 mt-3 pt-2 border-t border-line">
                  <button
                    onClick={() => openStatusModal(u)}
                    className="flex-1 text-[12px] font-semibold text-ink-3 border border-line rounded-lg py-1.5 active:bg-paper-2 transition-colors"
                  >
                    {u.status === 'active' ? 'Suspend' : 'Aktifkan'}
                  </button>
                  <button
                    onClick={() => openRoleModal(u)}
                    className="flex-1 text-[12px] font-semibold text-accent border border-accent/30 rounded-lg py-1.5 active:bg-accent-soft transition-colors"
                  >
                    Ubah Role
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-3 mt-6">
            <button
              disabled={page <= 1}
              onClick={() => { setPage(p => p - 1); load(page - 1); }}
              className="px-4 py-2 text-[12px] font-semibold border border-line rounded-xl disabled:opacity-40 active:bg-paper-2"
            >
              ← Prev
            </button>
            <span className="px-4 py-2 text-[12px] text-ink-3">
              {page} / {totalPages}
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => { setPage(p => p + 1); load(page + 1); }}
              className="px-4 py-2 text-[12px] font-semibold border border-line rounded-xl disabled:opacity-40 active:bg-paper-2"
            >
              Next →
            </button>
          </div>
        )}
      </div>

      {/* Admin nav links */}
      <div className="px-4 pb-6">
        <Link
          href="/admin/applications"
          className="glass flex items-center justify-between px-4 py-3 active:scale-[0.99] transition-transform"
        >
          <div>
            <p className="text-[13px] font-semibold text-ink">Aplikasi Collector</p>
            <p className="text-[11px] text-ink-4">Review pendaftaran collector baru</p>
          </div>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-ink-4 shrink-0">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </Link>
      </div>

      {/* ── Modals ── */}

      {/* Status Modal */}
      {modal?.type === 'status' && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/40" onClick={() => setModal(null)}>
          <div className="bg-surface w-full max-w-md rounded-t-2xl px-6 pt-5 pb-8" onClick={e => e.stopPropagation()}>
            <p className="text-[15px] font-bold text-ink mb-1">Ubah Status Pengguna</p>
            <p className="text-[12px] text-ink-3 mb-4">{modal.user.full_name}</p>

            {actionErr && <p className="text-[12px] text-red-600 mb-3">{actionErr}</p>}

            <div className="space-y-2 mb-4">
              {(['active', 'suspended'] as UserStatus[]).map(s => (
                <button
                  key={s}
                  onClick={() => setNewStatus(s)}
                  className={`w-full text-left px-4 py-3 rounded-xl text-[13px] font-semibold border transition-colors ${
                    newStatus === s ? 'border-accent bg-accent-soft text-accent' : 'border-line text-ink-3'
                  }`}
                >
                  {USER_STATUS_LABEL[s]}
                </button>
              ))}
            </div>

            <textarea
              placeholder="Alasan (opsional)"
              value={statusReason}
              onChange={e => setStatusReason(e.target.value)}
              rows={2}
              className="w-full border border-line rounded-xl px-4 py-3 text-[13px] focus:outline-none focus:border-accent resize-none mb-4"
            />

            <div className="flex gap-3">
              <button onClick={() => setModal(null)} className="flex-1 py-3 border border-line rounded-xl text-[13px] font-semibold text-ink-3">
                Batal
              </button>
              <button
                disabled={busy}
                onClick={submitStatus}
                className="flex-1 py-3 bg-accent text-white rounded-xl text-[13px] font-semibold disabled:opacity-50"
              >
                {busy ? 'Menyimpan...' : 'Simpan'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Role Modal */}
      {modal?.type === 'role' && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/40" onClick={() => setModal(null)}>
          <div className="bg-surface w-full max-w-md rounded-t-2xl px-6 pt-5 pb-8" onClick={e => e.stopPropagation()}>
            <p className="text-[15px] font-bold text-ink mb-1">Ubah Role Pengguna</p>
            <p className="text-[12px] text-ink-3 mb-4">{modal.user.full_name}</p>

            {actionErr && <p className="text-[12px] text-red-600 mb-3">{actionErr}</p>}

            <div className="space-y-2 mb-4">
              {(['user', 'cv', 'collector', 'admin', 'super_admin'] as UserRole[]).map(r => (
                <button
                  key={r}
                  onClick={() => setNewRole(r)}
                  className={`w-full text-left px-4 py-3 rounded-xl text-[13px] font-semibold border transition-colors ${
                    newRole === r ? 'border-accent bg-accent-soft text-accent' : 'border-line text-ink-3'
                  }`}
                >
                  {USER_ROLE_LABEL[r]}
                </button>
              ))}
            </div>

            <textarea
              placeholder="Alasan perubahan (opsional)"
              value={roleReason}
              onChange={e => setRoleReason(e.target.value)}
              rows={2}
              className="w-full border border-line rounded-xl px-4 py-3 text-[13px] focus:outline-none focus:border-accent resize-none mb-4"
            />

            <div className="flex gap-3">
              <button onClick={() => setModal(null)} className="flex-1 py-3 border border-line rounded-xl text-[13px] font-semibold text-ink-3">
                Batal
              </button>
              <button
                disabled={busy}
                onClick={submitRole}
                className="flex-1 py-3 bg-accent text-white rounded-xl text-[13px] font-semibold disabled:opacity-50"
              >
                {busy ? 'Menyimpan...' : 'Simpan'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
