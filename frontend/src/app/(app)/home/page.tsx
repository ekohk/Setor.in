import { auth } from '@/lib/auth';
import { api } from '@/lib/api';
import Link from 'next/link';
import type { Order } from '@/types/api';
import { ORDER_STATUS_LABEL, isActiveOrder } from '@/types/api';

// ─── Helpers ────────────────────────────────────────────────────────────────

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(amount);
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 60) return `${mins} menit lalu`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs} jam lalu`;
  return `${Math.floor(hrs / 24)} hari lalu`;
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

// ─── Server component ────────────────────────────────────────────────────────

export default async function HomePage() {
  const session = await auth();
  const userName = session?.user?.name?.split(' ')[0] ?? 'Pengguna';

  // Fetch orders (may fail if backend not running — we handle gracefully)
  let activeOrders: Order[] = [];
  let totalDone = 0;
  let totalEarned = 0;

  try {
    // api() returns envelope.data — for paginated endpoints that's Order[]
    const orders = await api<Order[]>('/v1/orders/me?page=1&page_size=50');
    activeOrders = orders.filter(o => isActiveOrder(o.status));
    const done = orders.filter(o => o.status === 'done');
    totalDone = done.length;
    totalEarned = done.reduce((sum, o) => sum + (o.final_payout ?? o.estimated_payout), 0);
  } catch {
    // Backend offline — show empty state gracefully
  }

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-8">
        <p className="text-accent-soft/70 text-sm font-medium">Selamat datang,</p>
        <h1 className="text-white text-2xl font-bold tracking-tight mt-0.5">{userName} 👋</h1>

        {/* Stats row */}
        <div className="mt-5 grid grid-cols-2 gap-3">
          <div className="bg-white/10 rounded-sm px-4 py-3">
            <p className="text-white/60 text-xs font-medium">Total Selesai</p>
            <p className="text-white text-xl font-bold mt-0.5">{totalDone} order</p>
          </div>
          <div className="bg-white/10 rounded-sm px-4 py-3">
            <p className="text-white/60 text-xs font-medium">Total Penghasilan</p>
            <p className="text-white text-xl font-bold mt-0.5">{formatRupiah(totalEarned)}</p>
          </div>
        </div>
      </div>

      <div className="px-4 py-5 space-y-5">
        {/* Quick actions */}
        <div className="grid grid-cols-2 gap-3">
          <Link
            href="/sell"
            className="glass flex items-center gap-3 px-4 py-4 active:scale-95 transition-transform"
          >
            <div className="w-10 h-10 rounded-sm bg-accent-soft flex items-center justify-center shrink-0">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" />
              </svg>
            </div>
            <div>
              <p className="text-[13px] font-bold text-ink">Jual Sampah</p>
              <p className="text-[11px] text-ink-3 mt-0.5">Buat order baru</p>
            </div>
          </Link>

          <Link
            href="/tracking"
            className="glass flex items-center gap-3 px-4 py-4 active:scale-95 transition-transform"
          >
            <div className="w-10 h-10 rounded-sm bg-accent-soft flex items-center justify-center shrink-0">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
              </svg>
            </div>
            <div>
              <p className="text-[13px] font-bold text-ink">Tracking</p>
              <p className="text-[11px] text-ink-3 mt-0.5">Lihat semua order</p>
            </div>
          </Link>
        </div>

        {/* Active orders */}
        <section>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-[15px] font-bold text-ink">Order Aktif</h2>
            <Link href="/tracking" className="text-[12px] font-semibold text-accent">Lihat semua</Link>
          </div>

          {activeOrders.length === 0 ? (
            <div className="glass flex flex-col items-center py-10 text-center">
              <div className="w-14 h-14 rounded-full bg-accent-soft flex items-center justify-center mb-3">
                <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="2" y="3" width="20" height="14" rx="2" /><line x1="8" y1="21" x2="16" y2="21" /><line x1="12" y1="17" x2="12" y2="21" />
                </svg>
              </div>
              <p className="text-[13px] font-semibold text-ink-2">Belum ada order aktif</p>
              <p className="text-[12px] text-ink-4 mt-1">Yuk jual sampah daur ulang kamu!</p>
              <Link href="/sell" className="mt-4 px-5 py-2 bg-accent text-white text-[13px] font-semibold rounded-sm">
                Mulai Jual
              </Link>
            </div>
          ) : (
            <div className="space-y-2">
              {activeOrders.slice(0, 3).map(order => (
                <Link key={order.id} href={`/tracking/${order.order_code}`} className="glass block px-4 py-3.5 active:scale-[0.99] transition-transform">
                  <div className="flex items-center justify-between">
                    <span className="text-[11px] font-bold text-ink-3">{order.order_code}</span>
                    <span className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${STATUS_COLOR[order.status] ?? 'bg-paper-2 text-ink-3'}`}>
                      {ORDER_STATUS_LABEL[order.status]}
                    </span>
                  </div>
                  <div className="flex items-center justify-between mt-2">
                    <p className="text-[13px] font-semibold text-ink capitalize">{order.method === 'pickup' ? 'Pickup' : 'Drop-off'}</p>
                    <p className="text-[13px] font-bold text-accent">{formatRupiah(order.estimated_payout)}</p>
                  </div>
                  <p className="text-[11px] text-ink-4 mt-1">{timeAgo(order.created_at)}</p>
                </Link>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
