import { Suspense } from 'react';
import Link from 'next/link';
import { BrandMark } from '@/components/ui/BrandMark';
import { LoginForm } from './LoginForm';

/**
 * /login — branded login page.
 *
 * Wraps Keycloak's Direct Access Grant via next-auth's Credentials provider.
 * UI matches the EcoCycle design system (project/styles.css). Mobile-first.
 *
 * URL params:
 *   ?error=... → next-auth surfaces sign-in errors via this query param
 *   ?callbackUrl=... → where to send the user after success (default /home)
 *   ?mode=register → switch CTA to register flow (TODO: wire up)
 */
export default function LoginPage() {
  return (
    <main className="min-h-screen relative">
      <div className="bg-mesh absolute inset-0 -z-10" />

      <div className="max-w-md mx-auto px-6 pt-12 pb-10 flex flex-col min-h-screen">
        {/* Header — back to landing + brand */}
        <div className="flex items-center justify-between mb-8">
          <Link
            href="/"
            className="w-10 h-10 grid place-items-center rounded-full bg-surface border border-line text-ink-2 hover:bg-paper-2 transition"
            aria-label="Kembali"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M19 12H5M11 6l-6 6 6 6" />
            </svg>
          </Link>
          <div className="flex items-center gap-2">
            <BrandMark size={26} />
            <span className="text-[15px] font-bold tracking-tight">Setor.in</span>
          </div>
          <div className="w-10" /> {/* spacer for symmetric header */}
        </div>

        {/* Title */}
        <div className="mb-8 animate-pageIn">
          <h1 className="text-[26px] font-bold tracking-tight text-ink mb-2">Selamat datang kembali</h1>
          <p className="text-[14px] text-ink-3 leading-relaxed">
            Masuk untuk mulai jual sampah Anda.
          </p>
        </div>

        {/* Form */}
        <div className="glass p-6 animate-pageIn">
          <Suspense fallback={<div className="h-[200px] grid place-items-center text-ink-3 text-sm">Memuat…</div>}>
            <LoginForm />
          </Suspense>
        </div>

        {/* Register CTA */}
        <div className="mt-8 text-center text-[13px] text-ink-3">
          Belum punya akun?{' '}
          <a
            href={
              (process.env.NEXT_PUBLIC_KEYCLOAK_BASE_URL ?? 'http://localhost:8090') +
              '/realms/' +
              (process.env.NEXT_PUBLIC_KEYCLOAK_REALM ?? 'setorin') +
              '/protocol/openid-connect/registrations?client_id=' +
              (process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT_ID ?? 'setorin-web') +
              '&response_type=code&redirect_uri=' +
              encodeURIComponent(
                (process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:3000') + '/login',
              )
            }
            className="text-accent-deep font-semibold underline-offset-2 hover:underline"
          >
            Daftar di sini
          </a>
        </div>

        {/* Footer */}
        <div className="mt-auto pt-10 text-center text-[11px] text-ink-4">
          © Setor.in · Marketplace daur ulang Indonesia
        </div>
      </div>
    </main>
  );
}
