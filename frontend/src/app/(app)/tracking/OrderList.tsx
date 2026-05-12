'use client';

import { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import type { Order, OrderStatus } from '@/types/api';
import { ORDER_STATUS_LABEL, isActiveOrder } from '@/types/api';
import { getMyOrders } from '@/lib/api/orders';

function formatRupiah(n: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n);
}

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(iso));
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

type FilterTab = 'active' | 'done' | 'all';

export default function OrderList() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<FilterTab>('active');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const result = await getMyOrders(1, 50);
      setOrders(result.items);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal memuat data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const filtered = orders.filter(o => {
    if (tab === 'active') return isActiveOrder(o.status as OrderStatus);
    if (tab === 'done') return o.status === 'done' || o.status === 'cancelled';
    return true;
  });

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-5">
        <h1 className="text-white text-2xl font-bold tracking-tight">Tracking Order</h1>
        <p className="text-white/70 text-sm mt-1">Pantau status sampah yang dijual</p>
      </div>

      {/* Filter tabs */}
      <div className="bg-surface border-b border-line px-4 flex gap-1 sticky top-0 z-10">
        {(['active', 'done', 'all'] as FilterTab[]).map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`py-3 px-4 text-[13px] font-semibold border-b-2 transition-colors ${
              tab === t
                ? 'border-accent text-accent'
                : 'border-transparent text-ink-3'
            }`}
          >
            {t === 'active' ? 'Aktif' : t === 'done' ? 'Selesai' : 'Semua'}
          </button>
        ))}
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

        {!loading && !error && filtered.length === 0 && (
          <div className="glass flex flex-col items-center py-12 text-center mt-2">
            <div className="w-14 h-14 rounded-full bg-accent-soft flex items-center justify-center mb-3">
              <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
              </svg>
            </div>
            <p className="text-[13px] font-semibold text-ink-2">Belum ada order {tab === 'active' ? 'aktif' : ''}</p>
            <Link href="/sell" className="mt-4 px-5 py-2 bg-accent text-white text-[13px] font-semibold rounded-sm">
              Buat Order Baru
            </Link>
          </div>
        )}

        {!loading && !error && filtered.length > 0 && (
          <div className="space-y-2">
            {filtered.map(order => (
              <Link
                key={order.id}
                href={`/tracking/${order.order_code}`}
                className="glass block px-4 py-4 active:scale-[0.99] transition-transform"
              >
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-[11px] font-bold text-ink-3">{order.order_code}</p>
                    <p className="text-[14px] font-bold text-ink mt-0.5 capitalize">
                      {order.method === 'pickup' ? 'Pickup' : 'Drop-off'}
                    </p>
                    <p className="text-[12px] text-ink-3 mt-0.5 line-clamp-1">{order.address_text}</p>
                  </div>
                  <div className="text-right shrink-0">
                    <span className={`text-[10px] font-semibold px-2 py-1 rounded-full ${STATUS_COLOR[order.status] ?? 'bg-paper-2 text-ink-3'}`}>
                      {ORDER_STATUS_LABEL[order.status as OrderStatus]}
                    </span>
                    <p className="text-[13px] font-bold text-accent mt-2">
                      {formatRupiah(order.final_payout ?? order.estimated_payout)}
                    </p>
                  </div>
                </div>
                <p className="text-[11px] text-ink-4 mt-2 border-t border-line pt-2">{formatDate(order.created_at)}</p>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
