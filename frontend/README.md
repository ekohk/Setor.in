# Setor.in Frontend

Next.js 15 PWA mobile-first untuk Setor.in marketplace daur ulang.

## Stack
- **Next.js 15** (App Router) + React 19 + TypeScript
- **Tailwind CSS** dengan design tokens dari [`../project/styles.css`](../project/styles.css)
- **Auth.js (next-auth v5)** dengan Credentials provider → wrap Keycloak Direct Access Grant
- **httpOnly session cookie** — token tidak pernah ada di client-side JS

## Quick Start

### 1. Install deps

```powershell
cd c:\Users\HP\OneDrive\Documents\Setor.in\frontend
npm install
```

(Pakai `pnpm` atau `yarn` boleh, tapi `npm` paling simpel.)

### 2. Setup .env.local

```powershell
copy .env.example .env.local
```

Edit `.env.local`:

- **`AUTH_SECRET`** — generate string random:
  ```powershell
  npx auth secret
  ```
  (atau pakai `openssl rand -base64 32` di Linux/Mac)

- Yang lain biarkan default (sudah cocok untuk dev lokal).

### 3. Pastikan backend & Keycloak jalan

```powershell
cd ..\backend
make up
go run ./cmd/api
```

### 4. Run frontend

```powershell
cd ..\frontend
npm run dev
```

Buka http://localhost:3000

## Halaman yang Sudah Tersedia (Sprint 0005 Iter 1)

| Path | Status | Auth |
|---|---|---|
| `/` | ✅ Landing branded | Public |
| `/login` | ✅ Login form (custom UI, Keycloak credential check) | Public |
| `/home` | 🚧 Placeholder (next iter) | Authed |
| `/sell` | 🚧 (next iter) | Authed |
| `/tracking` | 🚧 (next iter) | Authed |

## Test Login

1. Pastikan user `eko@test.local` / `Eko1234!` ada di Keycloak (sudah dari sprint sebelumnya)
2. Buka http://localhost:3000
3. Klik **Masuk ke akun**
4. Input email + password
5. Sukses → redirect ke `/home` (sekarang masih 404 karena belum dibuat — tapi cookie session sudah ter-set)

## Verify Login Berhasil

- Buka DevTools → Application → Cookies → `localhost:3000`
- Harus ada cookie `authjs.session-token` (httpOnly ✓)
- Token Keycloak ada di dalam cookie terenkripsi, **bukan di localStorage**

## Folder Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx              ← root layout, font, metadata
│   │   ├── page.tsx                ← / landing
│   │   ├── globals.css             ← Tailwind + design tokens
│   │   ├── (auth)/
│   │   │   └── login/
│   │   │       ├── page.tsx        ← /login (server component)
│   │   │       └── LoginForm.tsx   ← client form
│   │   └── api/auth/[...nextauth]/
│   │       └── route.ts            ← Auth.js handler
│   ├── components/
│   │   └── ui/
│   │       └── BrandMark.tsx       ← logo SVG
│   ├── lib/
│   │   ├── auth.ts                 ← Auth.js config (Credentials provider, JWT callbacks, refresh)
│   │   ├── auth-handlers.ts        ← re-export GET/POST
│   │   └── api.ts                  ← server-side fetch wrapper
│   ├── types/
│   │   └── next-auth.d.ts          ← TS module augmentation
│   └── middleware.ts               ← protect routes
├── tailwind.config.ts              ← design tokens
├── next.config.mjs
├── tsconfig.json
└── package.json
```

## Design System

Semua warna, radius, font size mengikuti [`../project/styles.css`](../project/styles.css):

- **Brand**: hijau `#2f7d52` (utama), `#1f5d3a` (deep), `#eaf3ec` (soft)
- **Ink**: `#131815` (primary text) → `#a3aaa3` (faint)
- **Surface**: `#ffffff` putih, `#f6f6f3` paper
- **Font**: Plus Jakarta Sans (loaded via Google Fonts)
- **Radius**: 18px default, 12px sm, 22px lg

Tailwind class: `bg-accent`, `text-accent-deep`, `bg-paper-2`, `text-ink-3`, `glass`, `bg-mesh`, dst.

## Auth Flow (Pattern B — BFF)

```
┌──────────┐  email+password  ┌────────────────┐
│  /login  │ ───────────────→ │ Auth.js route  │
│  (form)  │                   │ /api/auth/...  │
└──────────┘                   └────────┬───────┘
                                        │ Direct Access Grant
                                        ▼
                                 ┌──────────────┐
                                 │   Keycloak   │
                                 │ /token endpt │
                                 └──────┬───────┘
                                        │ access_token + refresh_token
                                        ▼
                                 ┌──────────────────────┐
                                 │ Auth.js sets httpOnly│
                                 │ encrypted JWT cookie │
                                 └──────────────────────┘
                                        │
                                        ▼
                              Browser only sees the cookie
                              (NEVER sees raw access_token)
```

## Production Notes

- Set `AUTH_URL` ke domain production
- Set `AUTH_SECRET` ke string random unik (jangan commit)
- Keycloak client `setorin-web` di Keycloak Admin → Settings → Web origins: tambah domain prod
- Pertimbangkan bundle dengan CSP header strict
