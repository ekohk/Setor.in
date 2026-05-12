'use client';

import { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import type { Order, OrderStatus } from '@/types/api';
import { ORDER_STATUS_LABEL, ORDER_STATUS_STEP } from '@/types/api';
import { cancelOrder, confirmCash } from '@/lib/api/orders';

function formatRupiah(n: number) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n);
}

function formatDateTime(iso: string) {
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  }).format(new Date(iso));
}

// Active stages for the timeline (terminal states excluded)
const TIMELINE_STAGES: OrderStatus[] = [
  'received', 'accepted', 'enroute', 'arrived',
  'weighing', 'quality', 'cash_handover', 'done',
];

function StatusTimeline({ status }: { status: OrderStatus }) {
  const currentStep = ORDER_STATUS_STEP[status];
  const isCancelled = status === 'cancelled' || status === 'disputed';

  if (isCancelled) {
    return (
      <div className="px-4 py-4">
        <div className="flex items-center gap-3 px-4 py-3 bg-red-50 rounded-sm border border-red-100">
          <span className="text-xl">❌</span>
          <div>
            <p className="text-[13px] font-bold text-red-700">
              {status === 'cancelled' ? 'Order Dibatalkan' : 'Order Dalam Sengketa'}
            </p>
            <p className="text-[12px] text-red-500 mt-0.5">
              {status === 'cancelled' ? 'Order ini telah dibatalkan' : 'Hubungi admin untuk penyelesaian'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="px-4 py-4">
      <p className="text-[12px] font-bold text-ink-3 mb-3 uppercase tracking-wider">Progress</p>
      <div className="space-y-0">
        {TIMELINE_STAGES.map((stage, idx) => {
          const step = ORDER_STATUS_STEP[stage];
          const isDone = step < currentStep;
          const isCurrent = step === currentStep;
          const isLast = idx === TIMELINE_STAGES.length - 1;

          return (
            <div key={stage} className="flex gap-3">
              {/* Dot + line */}
              <div className="flex flex-col items-center w-5 shrink-0">
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 ${
                  isDone ? 'bg-accent border-accent' : isCurrent ? 'bg-white border-accent' : 'bg-white border-line'
                }`}>
                  {isDone && (
                    <svg width="10" height="10" viewBox="0 0 12 12" fill="none">
                      <path d="M2 6l3 3 5-5" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  )}
                  {isCurrent && <div className="w-2 h-2 rounded-full bg-accent" />}
                </div>
                {!isLast && (
                  <div className={`w-0.5 flex-1 my-1 ${isDone ? 'bg-accent' : 'bg-line'}`} style={{ minHeight: 16 }} />
                )}
              </div>

              {/* Label */}
              <div className="pb-4">
                <p className={`text-[13px] font-semibold leading-tight ${
                  isCurrent ? 'text-accent' : isDone ? 'text-ink-3' : 'text-ink-4'
                }`}>
                  {ORDER_STATUS_LABEL[stage]}
                </p>
                {isCurrent && (
                  <p className="text-[11px] text-accent/70 mt-0.5">Status saat ini</p>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

interface Props {
  initialOrder: Order;
}

export default function OrderDetail({ initialOrder }: Props) {
  const router = useRouter();
  const [order, setOrder] = useState<Order>(initialOrder);
  const [isPending, startTransition] = useTransition();
  const [cancelReason, setCancelReason] = useState('');
  const [showCancelForm, setShowCancelForm] = useState(false);
  const [actionError, setActionError] = useState('');

  function handleCancel() {
    if (!cancelReason.trim() || cancelReason.trim().length < 3) {
      setActionError('Alasan pembatalan minimal 3 karakter.');
      return;
    }
    setActionError('');
    startTransition(async () => {
      try {
        const updated = await cancelOrder(order.order_code, cancelReason);
        setOrder(updated);
        setShowCancelForm(false);
      } catch (e) {
        setActionError(e instanceof Error ? e.message : 'Gagal membatalkan order');
      }
    });
  }

  function handleConfirmCash() {
    setActionError('');
    startTransition(async () => {
      try {
        const updated = await confirmCash(order.order_code);
        setOrder(updated);
      } catch (e) {
        setActionError(e instanceof Error ? e.message : 'Gagal konfirmasi');
      }
    });
  }

  const isTerminal = order.status === 'done' || order.status === 'cancelled' || order.status === 'disputed';

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-6">
        <button onClick={() => router.back()} className="flex items-center gap-1.5 text-white/80 text-[13px] mb-3">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
          Kembali
        </button>
        <p className="text-white/60 text-[11px] font-bold uppercase tracking-wider">Order</p>
        <h1 className="text-white text-xl font-bold tracking-tight mt-0.5">{order.order_code}</h1>
        <p className="text-white/70 text-[12px] mt-1">{formatDateTime(order.created_at)}</p>
      </div>

      {/* OTP banner — shown when accepted */}
      {order.status === 'accepted' && order.otp_code && (
        <div className="mx-4 mt-4 px-4 py-4 bg-amber-50 rounded-sm border border-amber-200">
          <p className="text-[12px] font-bold text-amber-700 uppercase tracking-wider">Kode OTP Anda</p>
          <p className="text-4xl font-bold text-amber-800 tracking-widest mt-1">{order.otp_code}</p>
          <p className="text-[11px] text-amber-600 mt-1">Tunjukkan ke driver saat tiba untuk verifikasi</p>
          {order.otp_expires_at && (
            <p className="text-[11px] text-amber-500 mt-0.5">Berlaku hingga {formatDateTime(order.otp_expires_at)}</p>
          )}
        </div>
      )}

      {/* Timeline */}
      <StatusTimeline status={order.status as OrderStatus} />

      {/* Details card */}
      <div className="px-4 pb-2">
        <div className="glass px-4 py-4 space-y-3">
          <p className="text-[12px] font-bold text-ink-3 uppercase tracking-wider">Rincian Order</p>

          <Row label="Metode" value={order.method === 'pickup' ? 'Pickup — Driver jemput' : 'Drop-off'} />
          <Row label="Estimasi Berat" value={`${order.estimated_weight_kg} kg`} />
          {order.actual_weight_kg && <Row label="Berat Aktual" value={`${order.actual_weight_kg} kg`} />}
          <Row label="Harga/kg" value={formatRupiah(order.unit_price_at_order)} />
          <Row label="Estimasi Payout" value={formatRupiah(order.estimated_payout)} highlight />
          {order.final_payout != null && <Row label="Payout Final" value={formatRupiah(order.final_payout)} highlight />}
          {order.quality_grade && <Row label="Grade Kualitas" value={`Grade ${order.quality_grade}`} />}
          <div className="pt-1 border-t border-line">
            <p className="text-[11px] text-ink-4">Alamat</p>
            <p className="text-[13px] text-ink mt-0.5">{order.address_text}</p>
          </div>
          {order.notes && (
            <div>
              <p className="text-[11px] text-ink-4">Catatan</p>
              <p className="text-[13px] text-ink mt-0.5">{order.notes}</p>
            </div>
          )}
          {order.cancellation_reason && (
            <div>
              <p className="text-[11px] text-ink-4">Alasan Pembatalan</p>
              <p className="text-[13px] text-red-600 mt-0.5">{order.cancellation_reason}</p>
            </div>
          )}
        </div>
      </div>

      {/* Action area */}
      {!isTerminal && (
        <div className="px-4 py-5 space-y-3">
          {/* Confirm cash button — only on cash_handover */}
          {order.status === 'cash_handover' && (
            <button
              onClick={handleConfirmCash}
              disabled={isPending}
              className="w-full py-3.5 bg-accent hover:bg-accent-deep disabled:opacity-60 text-white font-bold text-[14px] rounded-sm transition-all"
            >
              {isPending ? 'Memproses...' : 'Konfirmasi Terima Uang'}
            </button>
          )}

          {/* Cancel — only on received */}
          {order.status === 'received' && !showCancelForm && (
            <button
              onClick={() => setShowCancelForm(true)}
              className="w-full py-3 border border-red-200 text-red-600 font-semibold text-[13px] rounded-sm"
            >
              Batalkan Order
            </button>
          )}

          {showCancelForm && (
            <div className="glass px-4 py-4 space-y-3">
              <p className="text-[13px] font-bold text-ink">Alasan Pembatalan</p>
              <textarea
                rows={2}
                value={cancelReason}
                onChange={e => setCancelReason(e.target.value)}
                placeholder="Mengapa ingin membatalkan?"
                className="w-full px-3 py-2.5 rounded-sm border border-line bg-paper text-ink text-[13px] placeholder:text-ink-4 focus:outline-none focus:border-accent resize-none"
              />
              <div className="flex gap-2">
                <button
                  onClick={() => setShowCancelForm(false)}
                  className="flex-1 py-2.5 border border-line text-ink-2 text-[13px] font-semibold rounded-sm"
                >
                  Batal
                </button>
                <button
                  onClick={handleCancel}
                  disabled={isPending}
                  className="flex-1 py-2.5 bg-red-500 hover:bg-red-600 disabled:opacity-60 text-white text-[13px] font-bold rounded-sm"
                >
                  {isPending ? 'Memproses...' : 'Batalkan Order'}
                </button>
              </div>
            </div>
          )}

          {actionError && (
            <p className="text-[12px] text-red-600 px-1">{actionError}</p>
          )}
        </div>
      )}
    </div>
  );
}

function Row({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div className="flex items-center justify-between">
      <p className="text-[12px] text-ink-3">{label}</p>
      <p className={`text-[13px] font-semibold ${highlight ? 'text-accent' : 'text-ink'}`}>{value}</p>
    </div>
  );
}
