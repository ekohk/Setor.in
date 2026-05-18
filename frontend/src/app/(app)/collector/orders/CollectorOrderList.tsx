'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import type { Order, OrderStatus } from '@/types/api';
import { ORDER_STATUS_LABEL } from '@/types/api';
import { getIncomingOrders, getMyCollectorOrders } from '@/lib/api/collector';

function formatRupiah(n: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n);
}

function formatWeight(w?: string): string {
  if (!w) return '—';
  const n = parseFloat(w);
  return isNaN(n) ? w : n.toLocaleString('id-ID', { minimumFractionDigits: 0, maximumFractionDigits: 3 });
}

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }).format(new Date(iso));
}

const STATUS_COLOR: Record<string, string> = {
  received:      'bg-amber-50 text-amber-700',
  accepted:      'bg-blue-50 text-blue-700',
  enroute:       'bg-blue-50 text-blue-700',
  arrived:       'bg-purple-50 text-purple-700',
  weighing:      'bg-purple-50 text-purple-700',
  quality:       'bg-purple-50 text-purple-700',
  cash_handover: 'bg-accent-soft text-accent',
  done:          'bg-accent-soft text-accent-deep',
  cancelled:     'bg-red-50 text-red-600',
  disputed:      'bg-red-50 text-red-700',
};

type TabKey = 'incoming' | 'mine';

export default function CollectorOrderList() {
  const [tab, setTab] = useState<TabKey>('incoming');
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = tab === 'incoming'
        ? await getIncomingOrders(1, 30)
        : await getMyCollectorOrders(1, 30);
      setOrders(result.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal memuat data');
    } finally {
      setLoading(false);
    }
  }, [tab]);

  useEffect(() => { load(); }, [load]);

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-5">
        <h1 className="text-white text-2xl font-bold tracking-tight">Dashboard Collector</h1>
        <p className="text-white/70 text-sm mt-1">Kelola order pickup & drop-off</p>
      </div>

      {/* Tabs */}
      <div className="bg-surface border-b border-line px-4 flex gap-1 sticky top-0 z-10">
        {(['incoming', 'mine'] as TabKey[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`py-3 px-4 text-[13px] font-semibold border-b-2 transition-colors ${
              tab === t ? 'border-accent text-accent' : 'border-transparent text-ink-3'
            }`}
          >
            {t === 'incoming' ? 'Order Masuk' : 'Order Saya'}
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
            <p className="mt-3 text-[13px]">Memuat order...</p>
          </div>
        )}

        {!loading && error && (
          <div className="glass flex flex-col items-center py-10 text-center">
            <p className="text-[13px] text-red-600">{error}</p>
            <button onClick={load} className="mt-3 text-[12px] font-semibold text-accent">Coba Lagi</button>
          </div>
        )}

        {!loading && !error && orders.length === 0 && (
          <div className="glass flex flex-col items-center py-12 text-center mt-2">
            <div className="w-14 h-14 rounded-full bg-accent-soft flex items-center justify-center mb-3">
              <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="1" y="3" width="15" height="13" /><polygon points="16 8 20 8 23 11 23 16 16 16 16 8" /><circle cx="5.5" cy="18.5" r="2.5" /><circle cx="18.5" cy="18.5" r="2.5" />
              </svg>
            </div>
            <p className="text-[13px] font-semibold text-ink-2">
              {tab === 'incoming' ? 'Tidak ada order masuk saat ini' : 'Belum ada order yang diambil'}
            </p>
            <p className="text-[12px] text-ink-4 mt-1">
              {tab === 'incoming' ? 'Cek lagi nanti atau refresh halaman' : 'Ambil order dari tab "Order Masuk"'}
            </p>
          </div>
        )}

        {!loading && !error && orders.length > 0 && (
          <div className="space-y-2">
            {orders.map(order => (
              <Link
                key={order.id}
                href={`/collector/orders/${order.order_code}`}
                className="glass block px-4 py-4 active:scale-[0.99] transition-transform"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-[11px] font-bold text-ink-3">{order.order_code}</p>
                    <p className="text-[14px] font-bold text-ink mt-0.5 capitalize">
                      {order.method === 'pickup' ? '🚛 Pickup' : '📍 Drop-off'}
                    </p>
                    <p className="text-[12px] text-ink-3 mt-0.5 line-clamp-1">{order.address_text}</p>
                  </div>
                  <div className="text-right shrink-0">
                    <span className={`text-[10px] font-semibold px-2 py-1 rounded-full ${STATUS_COLOR[order.status] ?? 'bg-paper-2 text-ink-3'}`}>
                      {ORDER_STATUS_LABEL[order.status as OrderStatus]}
                    </span>
                    <p className="text-[13px] font-bold text-accent mt-2">
                      ~{formatRupiah(order.estimated_payout)}
                    </p>
                    <p className="text-[11px] text-ink-4 mt-0.5">{formatWeight(order.estimated_weight_kg)} kg</p>
                  </div>
                </div>
                <p className="text-[11px] text-ink-4 mt-2 border-t border-line pt-2">
                  {formatDate(order.created_at)}
                </p>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
