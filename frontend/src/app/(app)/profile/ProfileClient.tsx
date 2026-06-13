'use client';

import { useState, useTransition } from 'react';
import { signOut } from 'next-auth/react';
import type { UserProfile } from '@/types/api';

// ─── helpers ─────────────────────────────────────────────────────────────────

function formatDate(iso: string) {
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }).format(new Date(iso));
}

const STATUS_LABEL: Record<string, string> = {
  active:               'Aktif',
  pending_verification: 'Menunggu Verifikasi',
  suspended:            'Ditangguhkan',
  deleted:              'Dihapus',
};

const ROLE_LABEL: Record<string, string> = {
  user:        'Pengguna',
  cv:          'CV Partner',
  collector:   'Collector',
  admin:       'Admin',
  super_admin: 'Super Admin',
};

// ─── Edit form ────────────────────────────────────────────────────────────────

interface EditFormProps {
  profile: UserProfile;
  onSave: (updated: UserProfile) => void;
  onCancel: () => void;
}

function EditForm({ profile, onSave, onCancel }: EditFormProps) {
  const [isPending, startTransition] = useTransition();
  const [fullName, setFullName] = useState(profile.full_name);
  const [phone, setPhone] = useState(profile.phone ?? '');
  const [error, setError] = useState('');

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');

    if (fullName.trim().length < 1) { setError('Nama tidak boleh kosong.'); return; }
    if (phone && !/^\+?[0-9]{8,15}$/.test(phone.replace(/\s/g, ''))) {
      setError('Format nomor telepon tidak valid.'); return;
    }

    startTransition(async () => {
      try {
        const payload: Record<string, string> = {};
        if (fullName.trim() !== profile.full_name) payload.full_name = fullName.trim();
        if ((phone.trim() || undefined) !== profile.phone) payload.phone = phone.trim() || '';

        if (Object.keys(payload).length === 0) { onCancel(); return; }

        const res = await fetch('/api/proxy/me', {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          cache: 'no-store',
        });
        const envelope = await res.json() as { data?: UserProfile; error?: { message: string } };
        if (!res.ok) throw new Error(envelope.error?.message ?? 'Gagal menyimpan perubahan');
        onSave(envelope.data!);
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Terjadi kesalahan');
      }
    });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 px-4 py-5">
      <div>
        <label className="block text-[13px] font-bold text-ink mb-1.5">Nama Lengkap</label>
        <input
          type="text"
          value={fullName}
          onChange={e => setFullName(e.target.value)}
          className="w-full px-4 py-3 rounded-sm border border-line bg-surface text-ink text-[14px] placeholder:text-ink-4 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30"
        />
      </div>
      <div>
        <label className="block text-[13px] font-bold text-ink mb-1.5">
          Nomor Telepon <span className="text-ink-4 font-normal">(opsional)</span>
        </label>
        <input
          type="tel"
          value={phone}
          onChange={e => setPhone(e.target.value)}
          placeholder="+628xxxxxxxxxx"
          className="w-full px-4 py-3 rounded-sm border border-line bg-surface text-ink text-[14px] placeholder:text-ink-4 focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30"
        />
      </div>

      {error && <p className="text-[12px] text-red-600 px-1">{error}</p>}

      <div className="flex gap-2 pt-1">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 py-3 border border-line text-ink-2 text-[13px] font-semibold rounded-sm"
        >
          Batal
        </button>
        <button
          type="submit"
          disabled={isPending}
          className="flex-1 py-3 bg-accent hover:bg-accent-deep disabled:opacity-60 text-white text-[13px] font-bold rounded-sm transition-colors"
        >
          {isPending ? 'Menyimpan...' : 'Simpan'}
        </button>
      </div>
    </form>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

interface Props {
  initialProfile: UserProfile;
}

export default function ProfileClient({ initialProfile }: Props) {
  const [profile, setProfile] = useState<UserProfile>(initialProfile);
  const [editing, setEditing] = useState(false);
  const [showLogout, setShowLogout] = useState(false);
  const [isPending, startTransition] = useTransition();

  const initials = profile.full_name
    .split(' ')
    .slice(0, 2)
    .map(w => w[0]?.toUpperCase() ?? '')
    .join('');

  function handleLogout() {
    startTransition(async () => {
      await signOut({ callbackUrl: '/login' });
    });
  }

  if (editing) {
    return (
      <div className="animate-pageIn">
        <div className="bg-accent px-6 pt-12 pb-6">
          <button
            onClick={() => setEditing(false)}
            className="flex items-center gap-1.5 text-white/80 text-[13px] mb-3"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="15 18 9 12 15 6" />
            </svg>
            Kembali
          </button>
          <h1 className="text-white text-2xl font-bold tracking-tight">Edit Profil</h1>
        </div>
        <EditForm
          profile={profile}
          onSave={updated => { setProfile(updated); setEditing(false); }}
          onCancel={() => setEditing(false)}
        />
      </div>
    );
  }

  return (
    <div className="animate-pageIn">
      {/* Header */}
      <div className="bg-accent px-6 pt-12 pb-8">
        <h1 className="text-white text-2xl font-bold tracking-tight">Profil</h1>

        {/* Avatar + name */}
        <div className="flex items-center gap-4 mt-5">
          <div className="w-16 h-16 rounded-full bg-white/20 flex items-center justify-center shrink-0">
            {profile.avatar_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={profile.avatar_url} alt={profile.full_name} className="w-full h-full rounded-full object-cover" />
            ) : (
              <span className="text-white text-xl font-bold">{initials}</span>
            )}
          </div>
          <div className="min-w-0">
            <p className="text-white text-lg font-bold truncate">{profile.full_name}</p>
            <p className="text-white/70 text-[13px] truncate">{profile.email}</p>
            <span className={`inline-block mt-1 text-[10px] font-semibold px-2 py-0.5 rounded-full ${
              profile.status === 'active' ? 'bg-white/20 text-white' : 'bg-red-400/30 text-red-100'
            }`}>
              {STATUS_LABEL[profile.status] ?? profile.status}
            </span>
          </div>
        </div>
      </div>

      <div className="px-4 py-5 space-y-4">
        {/* Info card */}
        <div className="glass px-4 py-4 space-y-3">
          <p className="text-[12px] font-bold text-ink-3 uppercase tracking-wider">Informasi Akun</p>

          <InfoRow label="Email" value={profile.email} />
          <InfoRow label="Nomor Telepon" value={profile.phone ?? '—'} />
          <InfoRow label="Peran" value={ROLE_LABEL[profile.primary_role] ?? profile.primary_role} />
          <InfoRow
            label="Verifikasi Email"
            value={profile.email_verified ? 'Terverifikasi' : 'Belum Terverifikasi'}
            valueClass={profile.email_verified ? 'text-accent font-semibold' : 'text-amber-600 font-semibold'}
          />
          {profile.last_login_at && (
            <InfoRow label="Login Terakhir" value={formatDate(profile.last_login_at)} />
          )}
          <InfoRow label="Bergabung" value={formatDate(profile.created_at)} />
        </div>

        {/* Edit button */}
        <button
          onClick={() => setEditing(true)}
          className="w-full glass flex items-center justify-between px-4 py-3.5 active:scale-[0.99] transition-transform"
        >
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-sm bg-accent-soft flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#2f7d52" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </div>
            <span className="text-[13px] font-semibold text-ink">Edit Profil</span>
          </div>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#a3aaa3" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </button>

        {/* Logout */}
        {!showLogout ? (
          <button
            onClick={() => setShowLogout(true)}
            className="w-full flex items-center gap-3 px-4 py-3.5 rounded-sm border border-red-100 bg-red-50 active:scale-[0.99] transition-transform"
          >
            <div className="w-9 h-9 rounded-sm bg-red-100 flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                <polyline points="16 17 21 12 16 7" />
                <line x1="21" y1="12" x2="9" y2="12" />
              </svg>
            </div>
            <span className="text-[13px] font-semibold text-red-600">Keluar</span>
          </button>
        ) : (
          <div className="glass px-4 py-4 space-y-3">
            <p className="text-[13px] font-bold text-ink">Yakin ingin keluar?</p>
            <p className="text-[12px] text-ink-3">Anda harus login kembali untuk menggunakan aplikasi.</p>
            <div className="flex gap-2">
              <button
                onClick={() => setShowLogout(false)}
                className="flex-1 py-2.5 border border-line text-ink-2 text-[13px] font-semibold rounded-sm"
              >
                Batal
              </button>
              <button
                onClick={handleLogout}
                disabled={isPending}
                className="flex-1 py-2.5 bg-red-500 hover:bg-red-600 disabled:opacity-60 text-white text-[13px] font-bold rounded-sm"
              >
                {isPending ? 'Keluar...' : 'Ya, Keluar'}
              </button>
            </div>
          </div>
        )}

        {/* App version */}
        <p className="text-center text-[11px] text-ink-4 pt-2">Setor.in v1.0.0 · MVP</p>
      </div>
    </div>
  );
}

function InfoRow({ label, value, valueClass }: { label: string; value: string; valueClass?: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <p className="text-[12px] text-ink-3 shrink-0">{label}</p>
      <p className={`text-[13px] text-right truncate ${valueClass ?? 'text-ink font-medium'}`}>{value}</p>
    </div>
  );
}
