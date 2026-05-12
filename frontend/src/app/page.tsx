import Link from 'next/link';
import { BrandMark } from '@/components/ui/BrandMark';

/**
 * Landing page — public, unauthenticated.
 *
 * Mobile-first: vertical hero, single CTA. Mirrors the visual language of the
 * EcoCycle Home screen (mesh backdrop, leaf accent, glass cards).
 */
export default function LandingPage() {
  return (
    <main className="min-h-screen relative">
      <div className="bg-mesh absolute inset-0 -z-10" />

      <div className="max-w-md mx-auto px-6 pt-16 pb-10 flex flex-col min-h-screen">
        {/* Brand header */}
        <div className="flex items-center gap-2.5">
          <BrandMark size={32} />
          <span className="text-[18px] font-bold tracking-tight">Setor.in</span>
        </div>

        {/* Hero */}
        <div className="flex-1 flex flex-col justify-center py-12">
          <div className="inline-flex w-fit px-3 py-1 rounded-full bg-accent-soft text-accent-deep text-[11px] font-bold uppercase tracking-wider mb-5">
            Marketplace daur ulang
          </div>

          <h1 className="text-[32px] leading-[1.15] font-bold tracking-tight text-ink mb-4">
            Ubah sampah jadi
            <br />
            <span className="text-accent-deep">penghasilan</span>.
          </h1>
          <p className="text-[15px] leading-[1.6] text-ink-3">
            Plastik, logam, kertas, dan lainnya — pickup di rumah atau drop-off ke collector
            terdekat. Cash diterima langsung.
          </p>

          {/* Feature pills */}
          <div className="mt-8 grid grid-cols-3 gap-2">
            {[
              { v: '8', l: 'Material' },
              { v: '24/7', l: 'Layanan' },
              { v: 'Cash', l: 'Pembayaran' },
            ].map((s) => (
              <div key={s.l} className="glass px-3 py-3 text-center">
                <div className="text-[18px] font-bold text-accent-deep tracking-tight">{s.v}</div>
                <div className="text-[10px] uppercase tracking-wider text-ink-3 font-semibold mt-0.5">
                  {s.l}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* CTA */}
        <div className="space-y-3">
          <Link
            href="/login"
            className="flex items-center justify-center gap-2 w-full h-[54px] rounded-2xl bg-accent text-white font-bold text-[14px] tracking-tight transition active:scale-[0.98] hover:bg-accent-deep"
          >
            Masuk ke akun
          </Link>
          <p className="text-center text-[12px] text-ink-3">
            Belum punya akun?{' '}
            <Link href="/login?mode=register" className="text-accent-deep font-semibold">
              Daftar gratis
            </Link>
          </p>
        </div>
      </div>
    </main>
  );
}
