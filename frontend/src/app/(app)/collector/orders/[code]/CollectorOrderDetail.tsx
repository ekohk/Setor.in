'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import type { Order, OrderStatus } from '@/types/api';
import { ORDER_STATUS_LABEL } from '@/types/api';
import {
  acceptOrder,
  startPickup,
  arriveOrder,
  verifyOTP,
  weighOrder,
  qualityOrder,
} from '@/lib/api/collector';

function formatRupiah(n: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n);
}

// Backend stores weight as decimal string e.g. "1.000" (3dp, dot separator).
// parseFloat handles this correctly regardless of locale.
function formatWeight(w?: string): string {
  if (!w) return '—';
  const n = parseFloat(w);
  if (isNaN(n)) return w;
  // Show up to 3 decimal places, strip trailing zeros
  return n.toLocaleString('id-ID', { minimumFractionDigits: 0, maximumFractionDigits: 3 });
}

function formatDate(iso?: string) {
  if (!iso) return '—';
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(iso));
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

const STEPS: { key: OrderStatus; label: string }[] = [
  { key: 'received',      label: 'Order Masuk' },
  { key: 'accepted',      label: 'Diterima' },
  { key: 'enroute',       label: 'Berangkat' },
  { key: 'arrived',       label: 'Tiba' },
  { key: 'weighing',      label: 'Timbang' },
  { key: 'quality',       label: 'Kualitas' },
  { key: 'cash_handover', label: 'Serah Uang' },
  { key: 'done',          label: 'Selesai' },
];

const STEP_INDEX: Partial<Record<OrderStatus, number>> = Object.fromEntries(
  STEPS.map((s, i) => [s.key, i])
);

interface Props {
  order: Order;
}

export default function CollectorOrderDetail({ order: initial }: Props) {
  const router = useRouter();
  const [order, setOrder] = useState<Order>(initial);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');

  // Input states
  const [otp, setOtp] = useState('');
  const [weight, setWeight] = useState('');
  const [grade, setGrade] = useState<'A' | 'B' | 'C'>('A');
  const [gradeNotes, setGradeNotes] = useState('');

  async function run(action: () => Promise<Order>) {
    setBusy(true);
    setErr('');
    try {
      const updated = await action();
      setOrder(updated);
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Terjadi kesalahan');
    } finally {
      setBusy(false);
    }
  }

  const currentStep = STEP_INDEX[order.status] ?? -1;
  const isTerminal = ['done', 'cancelled', 'disputed'].includes(order.status);

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-5">
        <button
          onClick={() => router.back()}
          className="flex items-center gap-1.5 text-white/80 text-[13px] mb-3 active:text-white"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
          Kembali
        </button>
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-white/70 text-[11px] font-semibold tracking-widest uppercase">Order</p>
            <h1 className="text-white text-xl font-bold tracking-tight mt-0.5">{order.order_code}</h1>
            <p className="text-white/70 text-[13px] mt-1 capitalize">
              {order.method === 'pickup' ? '🚛 Pickup' : '📍 Drop-off'}
            </p>
          </div>
          <span className={`text-[11px] font-semibold px-2.5 py-1 rounded-full ${STATUS_COLOR[order.status] ?? 'bg-paper-2 text-ink-3'}`}>
            {ORDER_STATUS_LABEL[order.status as OrderStatus]}
          </span>
        </div>
      </div>

      {/* Progress bar */}
      {!isTerminal && (
        <div className="bg-surface px-4 py-3 border-b border-line overflow-x-auto">
          <div className="flex items-center gap-0 min-w-max">
            {STEPS.map((step, idx) => {
              const done = idx < currentStep;
              const active = idx === currentStep;
              return (
                <div key={step.key} className="flex items-center">
                  <div className="flex flex-col items-center gap-1">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold transition-colors ${
                      done ? 'bg-accent text-white' : active ? 'bg-accent text-white ring-2 ring-accent/30' : 'bg-paper-2 text-ink-4'
                    }`}>
                      {done ? (
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      ) : idx + 1}
                    </div>
                    <span className={`text-[9px] font-medium ${active ? 'text-accent' : done ? 'text-ink-3' : 'text-ink-4'}`}>
                      {step.label}
                    </span>
                  </div>
                  {idx < STEPS.length - 1 && (
                    <div className={`w-6 h-[2px] mb-4 ${done ? 'bg-accent' : 'bg-paper-2'}`} />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      <div className="px-4 py-4 space-y-3">
        {/* Error banner */}
        {err && (
          <div className="glass bg-red-50 border border-red-100 px-4 py-3">
            <p className="text-[13px] text-red-600">{err}</p>
          </div>
        )}

        {/* Order info */}
        <div className="glass px-4 py-4 space-y-3">
          <p className="text-[11px] font-bold text-ink-4 uppercase tracking-widest">Info Order</p>

          <div className="flex justify-between text-[13px]">
            <span className="text-ink-3">Alamat</span>
            <span className="text-ink font-medium text-right max-w-[60%]">{order.address_text}</span>
          </div>

          {order.notes && (
            <div className="flex justify-between text-[13px]">
              <span className="text-ink-3">Catatan</span>
              <span className="text-ink text-right max-w-[60%]">{order.notes}</span>
            </div>
          )}

          <div className="flex justify-between text-[13px]">
            <span className="text-ink-3">Est. Berat</span>
            <span className="text-ink font-medium">{formatWeight(order.estimated_weight_kg)} kg</span>
          </div>

          {order.actual_weight_kg && (
            <div className="flex justify-between text-[13px]">
              <span className="text-ink-3">Berat Aktual</span>
              <span className="text-ink font-bold text-accent">{formatWeight(order.actual_weight_kg)} kg</span>
            </div>
          )}

          <div className="flex justify-between text-[13px]">
            <span className="text-ink-3">Est. Payout</span>
            <span className="text-ink font-bold">~{formatRupiah(order.estimated_payout)}</span>
          </div>

          {order.final_payout !== undefined && order.final_payout !== null && (
            <div className="flex justify-between text-[13px]">
              <span className="text-ink-3">Final Payout</span>
              <span className="text-accent font-bold text-[15px]">{formatRupiah(order.final_payout)}</span>
            </div>
          )}

          {order.quality_grade && (
            <div className="flex justify-between text-[13px]">
              <span className="text-ink-3">Grade</span>
              <span className="text-ink font-bold">Grade {order.quality_grade}</span>
            </div>
          )}

          <div className="flex justify-between text-[13px]">
            <span className="text-ink-3">Dibuat</span>
            <span className="text-ink-3">{formatDate(order.created_at)}</span>
          </div>
        </div>

        {/* ── Action panel ── */}

        {/* received → Accept */}
        {order.status === 'received' && (
          <div className="glass px-4 py-4">
            <p className="text-[13px] font-semibold text-ink mb-3">Ambil order ini?</p>
            <button
              disabled={busy}
              onClick={() => run(() => acceptOrder(order.order_code))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Memproses...' : 'Terima Order'}
            </button>
          </div>
        )}

        {/* accepted + pickup → Start Pickup */}
        {order.status === 'accepted' && order.method === 'pickup' && (
          <div className="glass px-4 py-4">
            <p className="text-[13px] text-ink-3 mb-3">Order diterima. Mulai perjalanan menuju lokasi pickup.</p>
            <button
              disabled={busy}
              onClick={() => run(() => startPickup(order.order_code))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Memproses...' : 'Mulai Pickup'}
            </button>
          </div>
        )}

        {/* accepted + dropoff → Arrive directly */}
        {order.status === 'accepted' && order.method === 'dropoff' && (
          <div className="glass px-4 py-4">
            <p className="text-[13px] text-ink-3 mb-3">Konfirmasi bahwa pelanggan telah tiba di lokasi drop-off.</p>
            <button
              disabled={busy}
              onClick={() => run(() => arriveOrder(order.order_code))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Memproses...' : 'Konfirmasi Tiba'}
            </button>
          </div>
        )}

        {/* enroute → Arrive */}
        {order.status === 'enroute' && (
          <div className="glass px-4 py-4">
            <p className="text-[13px] text-ink-3 mb-3">Konfirmasi bahwa Anda sudah tiba di lokasi pelanggan.</p>
            <button
              disabled={busy}
              onClick={() => run(() => arriveOrder(order.order_code))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Memproses...' : 'Konfirmasi Tiba'}
            </button>
          </div>
        )}

        {/* arrived → Verify OTP */}
        {order.status === 'arrived' && (
          <div className="glass px-4 py-4 space-y-3">
            <p className="text-[13px] font-semibold text-ink">Verifikasi OTP Pelanggan</p>
            <p className="text-[12px] text-ink-3">Minta kode 4-digit dari pelanggan untuk memverifikasi kehadiran.</p>
            <input
              type="number"
              inputMode="numeric"
              maxLength={4}
              placeholder="Kode OTP (4 digit)"
              value={otp}
              onChange={e => setOtp(e.target.value.slice(0, 4))}
              className="w-full border border-line rounded-xl px-4 py-3 text-[15px] font-bold tracking-[0.5em] text-center focus:outline-none focus:border-accent"
            />
            <button
              disabled={busy || otp.length !== 4}
              onClick={() => run(() => verifyOTP(order.order_code, otp))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Memverifikasi...' : 'Verifikasi OTP'}
            </button>
          </div>
        )}

        {/* weighing → Enter actual weight */}
        {order.status === 'weighing' && (
          <div className="glass px-4 py-4 space-y-3">
            <p className="text-[13px] font-semibold text-ink">Input Berat Aktual</p>
            <div className="relative">
              <input
                type="number"
                inputMode="decimal"
                placeholder="0.0"
                value={weight}
                onChange={e => setWeight(e.target.value)}
                className="w-full border border-line rounded-xl px-4 py-3 pr-14 text-[15px] focus:outline-none focus:border-accent"
              />
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-[13px] text-ink-4 font-semibold">kg</span>
            </div>
            <button
              disabled={busy || !weight || parseFloat(weight) <= 0}
              onClick={() => run(() => weighOrder(order.order_code, weight))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Menyimpan...' : 'Simpan Berat'}
            </button>
          </div>
        )}

        {/* quality → Select grade */}
        {order.status === 'quality' && (
          <div className="glass px-4 py-4 space-y-3">
            <p className="text-[13px] font-semibold text-ink">Penilaian Kualitas</p>
            <div className="grid grid-cols-3 gap-2">
              {(['A', 'B', 'C'] as const).map(g => (
                <button
                  key={g}
                  onClick={() => setGrade(g)}
                  className={`py-3 rounded-xl text-[15px] font-bold transition-colors ${
                    grade === g ? 'bg-accent text-white' : 'bg-paper-2 text-ink-3'
                  }`}
                >
                  Grade {g}
                </button>
              ))}
            </div>
            <div className="text-[11px] text-ink-4 space-y-1 px-1">
              <p><span className="font-semibold text-accent">Grade A</span> — Bersih, kering <span className="text-accent font-bold">(harga penuh)</span></p>
              <p><span className="font-semibold text-blue-600">Grade B</span> — Sedikit kotor/basah <span className="text-blue-600 font-bold">(- Rp 1.000/kg)</span></p>
              <p><span className="font-semibold text-red-500">Grade C</span> — Kotor/terkontaminasi <span className="text-red-500 font-bold">(- Rp 2.000/kg)</span></p>
            </div>
            <textarea
              placeholder="Catatan kualitas (opsional)"
              value={gradeNotes}
              onChange={e => setGradeNotes(e.target.value)}
              rows={2}
              className="w-full border border-line rounded-xl px-4 py-3 text-[13px] focus:outline-none focus:border-accent resize-none"
            />
            <button
              disabled={busy}
              onClick={() => run(() => qualityOrder(order.order_code, grade, gradeNotes || undefined))}
              className="w-full bg-accent text-white font-semibold py-3 rounded-xl text-[14px] disabled:opacity-50 active:scale-[0.98] transition-transform"
            >
              {busy ? 'Menyimpan...' : 'Simpan Penilaian'}
            </button>
          </div>
        )}

        {/* cash_handover — waiting for user to confirm receipt */}
        {order.status === 'cash_handover' && (
          <div className="glass px-4 py-5 flex flex-col items-center text-center gap-3">
            <div className="w-14 h-14 rounded-full bg-accent-soft flex items-center justify-center">
              <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <line x1="12" y1="1" x2="12" y2="23" /><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <p className="text-[14px] font-bold text-ink">Menunggu Konfirmasi Pelanggan</p>
            <p className="text-[12px] text-ink-3 leading-relaxed">
              Serahkan uang sebesar{' '}
              <span className="font-bold text-accent">
                {order.final_payout !== undefined && order.final_payout !== null
                  ? formatRupiah(order.final_payout)
                  : `~${formatRupiah(order.estimated_payout)}`}
              </span>{' '}
              kepada pelanggan. Pelanggan akan mengkonfirmasi penerimaan dari aplikasi mereka.
            </p>
          </div>
        )}

        {/* done */}
        {order.status === 'done' && (
          <div className="glass px-4 py-5 flex flex-col items-center text-center gap-2">
            <div className="w-14 h-14 rounded-full bg-accent-soft flex items-center justify-center">
              <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>
            <p className="text-[15px] font-bold text-ink">Order Selesai</p>
            <p className="text-[12px] text-ink-3">
              Payout: <span className="font-bold text-accent">
                {order.final_payout !== undefined && order.final_payout !== null
                  ? formatRupiah(order.final_payout)
                  : '—'}
              </span>
            </p>
            {order.completed_at && (
              <p className="text-[11px] text-ink-4">{formatDate(order.completed_at)}</p>
            )}
          </div>
        )}

        {/* cancelled */}
        {order.status === 'cancelled' && (
          <div className="glass px-4 py-4 bg-red-50/50">
            <p className="text-[13px] font-semibold text-red-600">Order Dibatalkan</p>
            {order.cancellation_reason && (
              <p className="text-[12px] text-red-500 mt-1">{order.cancellation_reason}</p>
            )}
            {order.cancelled_at && (
              <p className="text-[11px] text-ink-4 mt-2">{formatDate(order.cancelled_at)}</p>
            )}
          </div>
        )}

        {/* disputed */}
        {order.status === 'disputed' && (
          <div className="glass px-4 py-4 bg-red-50/50">
            <p className="text-[13px] font-semibold text-red-700">Order Dalam Sengketa</p>
            <p className="text-[12px] text-red-500 mt-1">Hubungi tim support untuk resolusi lebih lanjut.</p>
          </div>
        )}
      </div>
    </div>
  );
}
