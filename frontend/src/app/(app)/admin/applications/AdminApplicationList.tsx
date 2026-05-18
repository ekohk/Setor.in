'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import type { CollectorApplication, ApplicationStatus } from '@/types/api';
import { APPLICATION_STATUS_LABEL } from '@/types/api';
import { getCollectorApplications, approveApplication, rejectApplication } from '@/lib/api/admin';

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(iso));
}

const STATUS_COLOR: Record<ApplicationStatus, string> = {
  pending:  'bg-amber-50 text-amber-700',
  approved: 'bg-accent-soft text-accent-deep',
  rejected: 'bg-red-50 text-red-600',
};

type TabKey = ApplicationStatus | 'all';

type Modal =
  | { type: 'approve'; app: CollectorApplication }
  | { type: 'reject';  app: CollectorApplication }
  | null;

export default function AdminApplicationList() {
  const [tab, setTab]         = useState<TabKey>('pending');
  const [apps, setApps]       = useState<CollectorApplication[]>([]);
  const [total, setTotal]     = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError]     = useState('');

  const [modal, setModal]       = useState<Modal>(null);
  const [busy, setBusy]         = useState(false);
  const [actionErr, setActionErr] = useState('');
  const [rejectReason, setRejectReason] = useState('');

  const PAGE_SIZE = 20;

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await getCollectorApplications({
        page: 1,
        page_size: PAGE_SIZE,
        status: tab === 'all' ? undefined : tab,
      });
      setApps(res.items);
      setTotal(res.total);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal memuat data');
    } finally {
      setLoading(false);
    }
  }, [tab]);

  useEffect(() => { load(); }, [load]);

  async function submitApprove() {
    if (!modal || modal.type !== 'approve') return;
    setBusy(true);
    setActionErr('');
    try {
      const updated = await approveApplication(modal.app.id);
      setApps(prev => prev.map(a => a.id === updated.id ? updated : a));
      setModal(null);
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal approve');
    } finally {
      setBusy(false);
    }
  }

  async function submitReject() {
    if (!modal || modal.type !== 'reject') return;
    if (!rejectReason.trim()) { setActionErr('Alasan penolakan wajib diisi'); return; }
    setBusy(true);
    setActionErr('');
    try {
      const updated = await rejectApplication(modal.app.id, rejectReason);
      setApps(prev => prev.map(a => a.id === updated.id ? updated : a));
      setModal(null);
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal reject');
    } finally {
      setBusy(false);
    }
  }

  const TABS: { key: TabKey; label: string }[] = [
    { key: 'pending',  label: 'Menunggu' },
    { key: 'approved', label: 'Disetujui' },
    { key: 'rejected', label: 'Ditolak' },
    { key: 'all',      label: 'Semua' },
  ];

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-5">
        <Link
          href="/admin/users"
          className="flex items-center gap-1.5 text-white/80 text-[13px] mb-3 active:text-white"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
          Manajemen User
        </Link>
        <h1 className="text-white text-2xl font-bold tracking-tight">Aplikasi Collector</h1>
        <p className="text-white/70 text-sm mt-1">{total} aplikasi ditemukan</p>
      </div>

      {/* Tabs */}
      <div className="bg-surface border-b border-line px-4 flex gap-1 sticky top-0 z-10">
        {TABS.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`py-3 px-3 text-[12px] font-semibold border-b-2 transition-colors ${
              tab === t.key ? 'border-accent text-accent' : 'border-transparent text-ink-3'
            }`}
          >
            {t.label}
          </button>
        ))}
        <button
          onClick={load}
          className="ml-auto py-3 px-2 text-ink-4 active:text-accent transition-colors"
          title="Refresh"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" />
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
          </svg>
        </button>
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
            <button onClick={load} className="mt-3 text-[12px] font-semibold text-accent">Coba Lagi</button>
          </div>
        )}

        {!loading && !error && apps.length === 0 && (
          <div className="glass flex flex-col items-center py-12 text-center mt-2">
            <p className="text-[13px] font-semibold text-ink-2">Tidak ada aplikasi</p>
          </div>
        )}

        {!loading && !error && apps.length > 0 && (
          <div className="space-y-3">
            {apps.map(app => (
              <div key={app.id} className="glass px-4 py-4">
                {/* Status badge */}
                <div className="flex items-start justify-between gap-2 mb-3">
                  <div className="min-w-0">
                    <p className="text-[14px] font-bold text-ink">{app.business_name}</p>
                    <p className="text-[11px] text-ink-4 mt-0.5">{formatDate(app.submitted_at)}</p>
                  </div>
                  <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-full shrink-0 ${STATUS_COLOR[app.status]}`}>
                    {APPLICATION_STATUS_LABEL[app.status]}
                  </span>
                </div>

                {/* Details */}
                <div className="space-y-1.5 text-[12px] border-t border-line pt-3">
                  <div className="flex gap-2">
                    <span className="text-ink-4 w-20 shrink-0">Alamat</span>
                    <span className="text-ink-3 line-clamp-2">{app.address}</span>
                  </div>
                  {app.license_no && (
                    <div className="flex gap-2">
                      <span className="text-ink-4 w-20 shrink-0">No. SIM</span>
                      <span className="text-ink-3">{app.license_no}</span>
                    </div>
                  )}
                  <div className="flex gap-2">
                    <span className="text-ink-4 w-20 shrink-0">KTP</span>
                    <a href={app.ktp_url} target="_blank" rel="noreferrer" className="text-accent underline truncate max-w-[60%]">
                      Lihat dokumen
                    </a>
                  </div>
                  {app.siup_url && (
                    <div className="flex gap-2">
                      <span className="text-ink-4 w-20 shrink-0">SIUP</span>
                      <a href={app.siup_url} target="_blank" rel="noreferrer" className="text-accent underline truncate max-w-[60%]">
                        Lihat dokumen
                      </a>
                    </div>
                  )}
                  {app.rejection_reason && (
                    <div className="flex gap-2">
                      <span className="text-ink-4 w-20 shrink-0">Alasan</span>
                      <span className="text-red-500">{app.rejection_reason}</span>
                    </div>
                  )}
                </div>

                {/* Actions — only for pending */}
                {app.status === 'pending' && (
                  <div className="flex gap-2 mt-3 pt-3 border-t border-line">
                    <button
                      onClick={() => { setActionErr(''); setModal({ type: 'reject', app }); setRejectReason(''); }}
                      className="flex-1 py-2 text-[12px] font-semibold text-red-600 border border-red-200 rounded-xl active:bg-red-50 transition-colors"
                    >
                      Tolak
                    </button>
                    <button
                      onClick={() => { setActionErr(''); setModal({ type: 'approve', app }); }}
                      className="flex-1 py-2 text-[12px] font-semibold text-white bg-accent rounded-xl active:opacity-90 transition-opacity"
                    >
                      Setujui
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── Approve Modal ── */}
      {modal?.type === 'approve' && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/40" onClick={() => setModal(null)}>
          <div className="bg-surface w-full max-w-md rounded-t-2xl px-6 pt-5 pb-8" onClick={e => e.stopPropagation()}>
            <p className="text-[15px] font-bold text-ink mb-1">Setujui Aplikasi?</p>
            <p className="text-[12px] text-ink-3 mb-4">
              <span className="font-semibold">{modal.app.business_name}</span> akan mendapatkan role Collector.
            </p>
            {actionErr && <p className="text-[12px] text-red-600 mb-3">{actionErr}</p>}
            <div className="flex gap-3">
              <button onClick={() => setModal(null)} className="flex-1 py-3 border border-line rounded-xl text-[13px] font-semibold text-ink-3">
                Batal
              </button>
              <button
                disabled={busy}
                onClick={submitApprove}
                className="flex-1 py-3 bg-accent text-white rounded-xl text-[13px] font-semibold disabled:opacity-50"
              >
                {busy ? 'Memproses...' : 'Ya, Setujui'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Reject Modal ── */}
      {modal?.type === 'reject' && (
        <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/40" onClick={() => setModal(null)}>
          <div className="bg-surface w-full max-w-md rounded-t-2xl px-6 pt-5 pb-8" onClick={e => e.stopPropagation()}>
            <p className="text-[15px] font-bold text-ink mb-1">Tolak Aplikasi</p>
            <p className="text-[12px] text-ink-3 mb-4">{modal.app.business_name}</p>
            {actionErr && <p className="text-[12px] text-red-600 mb-3">{actionErr}</p>}
            <textarea
              placeholder="Alasan penolakan (wajib diisi)..."
              value={rejectReason}
              onChange={e => setRejectReason(e.target.value)}
              rows={3}
              className="w-full border border-line rounded-xl px-4 py-3 text-[13px] focus:outline-none focus:border-red-400 resize-none mb-4"
            />
            <div className="flex gap-3">
              <button onClick={() => setModal(null)} className="flex-1 py-3 border border-line rounded-xl text-[13px] font-semibold text-ink-3">
                Batal
              </button>
              <button
                disabled={busy || !rejectReason.trim()}
                onClick={submitReject}
                className="flex-1 py-3 bg-red-500 text-white rounded-xl text-[13px] font-semibold disabled:opacity-50"
              >
                {busy ? 'Memproses...' : 'Tolak'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
