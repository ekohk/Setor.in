'use client';

import { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import type { Material } from '@/types/api';
import { createOrder } from '@/lib/api/orders';

// ─── Material icon map (slug → emoji fallback) ───────────────────────────────
const MATERIAL_EMOJI: Record<string, string> = {
  plastic:   '🧴',
  cardboard: '📦',
  paper:     '📄',
  aluminum:  '🥫',
  copper:    '🔩',
  steel:     '⚙️',
  glass:     '🫙',
  ewaste:    '💻',
};

function formatRupiah(amount: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(amount);
}

interface Props {
  materials: Material[];
}

export default function SellForm({ materials }: Props) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();

  const [selectedSlug, setSelectedSlug] = useState<string>(materials[0]?.slug ?? '');
  const [weight, setWeight] = useState('');
  const [method, setMethod] = useState<'pickup' | 'dropoff'>('pickup');
  const [address, setAddress] = useState('');
  const [notes, setNotes] = useState('');
  const [error, setError] = useState('');

  const selectedMaterial = materials.find(m => m.slug === selectedSlug);
  const weightNum = parseFloat(weight.replace(',', '.'));
  const estimatedPayout =
    selectedMaterial?.current_price && !isNaN(weightNum) && weightNum > 0
      ? Math.floor(selectedMaterial.current_price * weightNum)
      : null;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');

    if (!selectedSlug) { setError('Pilih jenis material terlebih dahulu.'); return; }
    if (!weight || isNaN(weightNum) || weightNum <= 0) { setError('Masukkan estimasi berat yang valid.'); return; }
    if (address.trim().length < 10) { setError('Alamat terlalu singkat (minimal 10 karakter).'); return; }

    startTransition(async () => {
      try {
        const order = await createOrder({
          material_slug: selectedSlug,
          estimated_weight_kg: weightNum.toFixed(3),
          method,
          address_text: address.trim(),
          notes: notes.trim() || undefined,
        });
        router.push(`/tracking/${order.order_code}`);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Gagal membuat order. Coba lagi.');
      }
    });
  }

  return (
    <form onSubmit={handleSubmit} className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-6">
        <h1 className="text-white text-2xl font-bold tracking-tight">Jual Sampah</h1>
        <p className="text-white/70 text-sm mt-1">Pilih material dan estimasi berat</p>
      </div>

      <div className="px-4 py-5 space-y-5">
        {/* Material picker */}
        <section>
          <label className="block text-[13px] font-bold text-ink mb-2.5">Jenis Material</label>
          <div className="grid grid-cols-2 gap-2">
            {materials.map(mat => {
              const active = mat.slug === selectedSlug;
              return (
                <button
                  key={mat.slug}
                  type="button"
                  onClick={() => setSelectedSlug(mat.slug)}
                  className={`flex items-center gap-3 px-3 py-3 rounded-sm border text-left transition-all ${
                    active
                      ? 'border-accent bg-accent-soft'
                      : 'border-line bg-surface'
                  }`}
                >
                  <span className="text-2xl">{MATERIAL_EMOJI[mat.slug] ?? '♻️'}</span>
                  <div className="min-w-0">
                    <p className={`text-[12px] font-bold truncate ${active ? 'text-accent-deep' : 'text-ink'}`}>{mat.name}</p>
                    <p className={`text-[11px] ${active ? 'text-accent' : 'text-ink-4'}`}>
                      {mat.current_price ? formatRupiah(mat.current_price) + '/kg' : '—'}
                    </p>
                  </div>
                </button>
              );
            })}
          </div>
        </section>

        {/* Weight */}
        <section>
          <label htmlFor="weight" className="block text-[13px] font-bold text-ink mb-2">
            Estimasi Berat (kg)
          </label>
          <input
            id="weight"
            type="number"
            inputMode="decimal"
            step="0.1"
            min="0.1"
            max="999"
            placeholder="Contoh: 3.5"
            value={weight}
            onChange={e => setWeight(e.target.value)}
            className="w-full px-4 py-3 rounded-sm border border-line bg-surface text-ink text-[14px] placeholder:text-ink-4 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30"
          />
          {estimatedPayout !== null && (
            <div className="mt-2 px-3 py-2 bg-accent-soft rounded-sm flex items-center justify-between">
              <span className="text-[12px] text-accent font-medium">Estimasi payout</span>
              <span className="text-[14px] font-bold text-accent-deep">{formatRupiah(estimatedPayout)}</span>
            </div>
          )}
        </section>

        {/* Method */}
        <section>
          <label className="block text-[13px] font-bold text-ink mb-2.5">Metode Pengambilan</label>
          <div className="grid grid-cols-2 gap-2">
            {(['pickup', 'dropoff'] as const).map(m => (
              <button
                key={m}
                type="button"
                onClick={() => setMethod(m)}
                className={`flex items-center gap-2.5 px-3 py-3 rounded-sm border text-left transition-all ${
                  method === m ? 'border-accent bg-accent-soft' : 'border-line bg-surface'
                }`}
              >
                <span className="text-xl">{m === 'pickup' ? '🚛' : '📍'}</span>
                <div>
                  <p className={`text-[12px] font-bold ${method === m ? 'text-accent-deep' : 'text-ink'}`}>
                    {m === 'pickup' ? 'Pickup' : 'Drop-off'}
                  </p>
                  <p className={`text-[11px] ${method === m ? 'text-accent' : 'text-ink-4'}`}>
                    {m === 'pickup' ? 'Driver jemput ke lokasi' : 'Antar ke titik drop'}
                  </p>
                </div>
              </button>
            ))}
          </div>
        </section>

        {/* Address */}
        <section>
          <label htmlFor="address" className="block text-[13px] font-bold text-ink mb-2">
            {method === 'pickup' ? 'Alamat Pickup' : 'Alamat Anda'}
          </label>
          <textarea
            id="address"
            rows={3}
            placeholder="Masukkan alamat lengkap..."
            value={address}
            onChange={e => setAddress(e.target.value)}
            className="w-full px-4 py-3 rounded-sm border border-line bg-surface text-ink text-[14px] placeholder:text-ink-4 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 resize-none"
          />
        </section>

        {/* Notes */}
        <section>
          <label htmlFor="notes" className="block text-[13px] font-bold text-ink mb-2">
            Catatan untuk Collector <span className="text-ink-4 font-normal">(opsional)</span>
          </label>
          <textarea
            id="notes"
            rows={2}
            placeholder="Contoh: barang di depan pagar, hubungi dulu sebelum datang"
            value={notes}
            onChange={e => setNotes(e.target.value)}
            className="w-full px-4 py-3 rounded-sm border border-line bg-surface text-ink text-[14px] placeholder:text-ink-4 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 resize-none"
          />
        </section>

        {/* Error */}
        {error && (
          <div className="px-4 py-3 bg-red-50 rounded-sm border border-red-100">
            <p className="text-[13px] text-red-600">{error}</p>
          </div>
        )}

        {/* Submit */}
        <button
          type="submit"
          disabled={isPending}
          className="w-full py-3.5 bg-accent hover:bg-accent-deep active:scale-[0.99] disabled:opacity-60 text-white font-bold text-[14px] rounded-sm transition-all"
        >
          {isPending ? 'Membuat Order...' : 'Buat Order'}
        </button>
      </div>
    </form>
  );
}
