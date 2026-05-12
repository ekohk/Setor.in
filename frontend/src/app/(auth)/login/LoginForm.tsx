'use client';

import { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { signIn } from 'next-auth/react';

/**
 * Client-side login form.
 *
 * Submits credentials to next-auth's `keycloak-credentials` provider, which
 * exchanges them for a Keycloak access_token via Direct Access Grant.
 * On success, next-auth sets an httpOnly session cookie; we redirect to /home.
 *
 * Surfaces these error states inline:
 *   - invalid_credentials
 *   - email_not_verified
 *   - account_suspended
 *   - network_error
 */
export function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const callbackUrl = params.get('callbackUrl') ?? '/home';

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(decodeError(params.get('error')));

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSubmitting(true);

    const res = await signIn('keycloak-credentials', {
      email,
      password,
      redirect: false,
    });

    setSubmitting(false);

    if (!res) {
      setError('Tidak bisa terhubung ke server. Coba lagi.');
      return;
    }
    if (res.error) {
      setError(decodeError(res.error) ?? 'Email atau password salah.');
      return;
    }
    router.replace(callbackUrl);
    router.refresh();
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      {/* Email */}
      <div>
        <label htmlFor="email" className="block text-[11px] font-bold text-ink-3 uppercase tracking-wider mb-2">
          Email
        </label>
        <input
          id="email"
          type="email"
          autoComplete="email"
          inputMode="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full h-[48px] px-4 rounded-xl bg-paper-2 border border-transparent focus:border-accent focus:bg-surface focus:outline-none text-[15px] text-ink placeholder:text-ink-4 transition"
          placeholder="kamu@email.com"
        />
      </div>

      {/* Password */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <label htmlFor="password" className="text-[11px] font-bold text-ink-3 uppercase tracking-wider">
            Password
          </label>
          <a
            href={
              (process.env.NEXT_PUBLIC_KEYCLOAK_BASE_URL ?? 'http://localhost:8090') +
              '/realms/' +
              (process.env.NEXT_PUBLIC_KEYCLOAK_REALM ?? 'setorin') +
              '/login-actions/reset-credentials?client_id=' +
              (process.env.NEXT_PUBLIC_KEYCLOAK_CLIENT_ID ?? 'setorin-web')
            }
            className="text-[12px] text-accent-deep font-semibold hover:underline"
          >
            Lupa?
          </a>
        </div>
        <div className="relative">
          <input
            id="password"
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full h-[48px] pl-4 pr-12 rounded-xl bg-paper-2 border border-transparent focus:border-accent focus:bg-surface focus:outline-none text-[15px] text-ink placeholder:text-ink-4 transition"
            placeholder="Min. 8 karakter"
          />
          <button
            type="button"
            onClick={() => setShowPassword((v) => !v)}
            className="absolute right-2 top-1/2 -translate-y-1/2 w-9 h-9 grid place-items-center text-ink-3 hover:text-ink-2 rounded-lg"
            aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
          >
            {showPassword ? <EyeOff /> : <Eye />}
          </button>
        </div>
      </div>

      {/* Inline error */}
      {error && (
        <div className="rounded-xl bg-[#fbe9e7] border-l-[3px] border-[#9c2a1f] px-4 py-3 text-[13px] text-ink leading-relaxed">
          {error}
        </div>
      )}

      {/* Submit */}
      <button
        type="submit"
        disabled={submitting || !email || !password}
        className="w-full h-[54px] rounded-2xl bg-accent text-white font-bold text-[14px] tracking-tight transition active:scale-[0.98] hover:bg-accent-deep disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center gap-2"
      >
        {submitting ? (
          <>
            <Spinner /> Masuk…
          </>
        ) : (
          <>Masuk</>
        )}
      </button>
    </form>
  );
}

// ─── Tiny inline icons (no external lib) ────────────────────────────────────

function Eye() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M2 12s4-8 10-8 10 8 10 8-4 8-10 8-10-8-10-8z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function EyeOff() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M17.94 17.94A10.94 10.94 0 0 1 12 20c-6 0-10-8-10-8a18.5 18.5 0 0 1 5.06-5.94" />
      <path d="M9.9 4.24A10.94 10.94 0 0 1 12 4c6 0 10 8 10 8a18.45 18.45 0 0 1-2.16 3.19" />
      <path d="M14.12 14.12a3 3 0 1 1-4.24-4.24" />
      <line x1="2" y1="2" x2="22" y2="22" />
    </svg>
  );
}

function Spinner() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="animate-spin">
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeOpacity="0.25" strokeWidth="3" />
      <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}

// ─── Error code → friendly message ───────────────────────────────────────────

function decodeError(code: string | null): string | null {
  if (!code) return null;
  switch (code) {
    case 'CredentialsSignin':
    case 'invalid_credentials':
      return 'Email atau password salah.';
    case 'invalid_grant':
      return 'Email atau password salah, atau email belum diverifikasi.';
    case 'email_not_verified':
      return 'Email belum diverifikasi. Cek inbox Anda untuk link verifikasi.';
    case 'account_suspended':
      return 'Akun Anda dinonaktifkan sementara. Hubungi support@setor.in.';
    case 'OAuthCallback':
    case 'Configuration':
      return 'Konfigurasi server bermasalah. Hubungi support@setor.in.';
    default:
      return null;
  }
}
