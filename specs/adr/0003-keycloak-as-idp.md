# ADR-0003: Keycloak as Identity Provider

## Status
Accepted — 2026-05-05

## Context
Setor.in butuh sistem auth yang:
- Mendukung email/password login + email verification + forgot password
- Multi-role (user, collector, admin, super_admin) dengan composite roles
- Aman (PII regulasi UU PDP Indonesia)
- Bisa scale ke OAuth/OIDC (untuk future SSO ke pihak ketiga)
- Tim backend kecil (1-2 orang) → tidak punya bandwidth maintain auth dari nol

Building auth dari nol punya risiko tinggi: password hashing salah, token
revocation lemah, brute force protection lemah, rate limiting lemah.

## Decision
**Pakai Keycloak 25 sebagai Identity Provider (IdP)**, backend Go sebagai
**Resource Server** (tidak pernah handle password).

### Skema
1. Frontend (Next.js) menggunakan `keycloak-js` adapter → user login langsung
   ke Keycloak, dapat JWT (access + refresh).
2. Frontend kirim `Authorization: Bearer <token>` ke backend Go.
3. Backend Go validate JWT pakai **JWKS** dari Keycloak
   (`/realms/setorin/protocol/openid-connect/certs`), cache 1 jam.
4. Backend extract `sub` (Keycloak user ID), `realm_access.roles` (roles),
   custom claim `db_user_id` (mapping ke local DB).
5. Local DB hanya simpan profil bisnis (full_name, phone, addresses, wallet).
   **Tidak pernah** simpan password.

### Realm setup
- Realm: `setorin`
- Clients:
  - `setorin-web` (public, PKCE, untuk frontend)
  - `setorin-backend` (confidential, service account untuk admin operations:
    create user, set role)
- Realm roles: `user`, `collector`, `admin`, `super_admin` (composite)
- Required actions: `VERIFY_EMAIL`, `UPDATE_PASSWORD` (saat reset)
- Email verification: WAJIB sebelum bisa login
- Password policy: min 8, 1 upper, 1 digit, 1 special

### Email
- Dev: MailHog (SMTP localhost:1025)
- Prod: Resend / SendGrid (TBD, lihat ADR mendatang)

## Consequences

### Positif
- Auth security best practice built-in (PKCE, refresh token rotation,
  brute force lockout, password policy)
- Email template & flow built-in (verify, reset, magic link future)
- Admin UI built-in untuk manage users & sessions
- OIDC compliance → siap integrate Google/Apple SSO nanti tinggal toggle
- Tim backend bisa fokus business logic, bukan auth plumbing

### Negatif
- **Tambah operational complexity**: 1 service tambahan untuk di-host & monitor
- **Learning curve curam**: realm/client/scope/mapper butuh 1-2 minggu paham
- **Resource heavy**: Keycloak butuh ~512MB RAM, JVM startup ~30-60 detik
- **Schema export/import** untuk versioning realm config tidak straightforward
- **Custom email template** kalau mau branding sendiri butuh theme override
  (tidak diprioritaskan di MVP — pakai default Keycloak)

### Mitigasi
- Versioning realm config: export `realm-export.json` ke git, import otomatis
  saat startup (lewat Docker `--import-realm`)
- Operational: pakai managed Keycloak (Cloud-IAM, Phase Two) di production
  kalau resource jadi masalah
- Custom email: skip MVP, pakai default. Phase 2 kalau perlu branding.

## Alternatives Considered

### A. Build sendiri (bcrypt + JWT + email)
- **Ditolak**: high security risk, tim kecil, banyak edge case (token rotation,
  rate limiting, lockout, password reset flow).

### B. Auth0 / Clerk (managed SaaS)
- **Ditolak**: cost. Auth0 free tier 7.000 MAU lalu $$$, Clerk $25/mo + $0.02
  per MAU. Setor.in target 100k+ MAU → biaya signifikan. Vendor lock-in tinggi.

### C. Supabase Auth
- **Ditolak**: bundling dengan Supabase DB. Kita pakai Postgres self-hosted.
  Pakai Supabase Auth saja artinya 2 IdP (kompleks).

### D. Ory Kratos + Hydra
- **Dipertimbangkan serius**. Lebih lightweight dari Keycloak, API-first,
  no admin UI built-in. **Ditolak** karena: tidak ada admin UI built-in
  (harus build sendiri), lebih sedikit dokumentasi & community.

### E. Firebase Auth
- **Ditolak**: vendor lock-in Google, kurang fleksibel untuk RBAC kompleks,
  data residency Indonesia tidak jelas.

## Migration Path Out
Kalau Keycloak jadi bottleneck:
- User data (`keycloak_id`, email) sudah ter-mirror di local DB → bisa migrate
  ke IdP lain dengan data export Keycloak + script ulang.
- Token validation di backend pakai JWKS standard → ganti IdP cukup ubah
  issuer URL & JWKS endpoint.

## Open Questions (akan dijawab saat implementasi)
- [ ] Token storage frontend: httpOnly cookie (lebih aman) atau localStorage
  (lebih mudah)? → Lihat `specs/0002-auth/security.md`
- [ ] Refresh token rotation strategy?
- [ ] Session management: server-side (Keycloak revoke) atau JWT-only?
