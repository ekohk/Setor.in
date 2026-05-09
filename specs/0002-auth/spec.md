# Spec 0002 — Authentication & User Management

**Status:** Draft → In Implementation
**Last updated:** 2026-05-05
**Owner:** @ekohk
**Depends on:** [adr/0003-keycloak-as-idp.md](../adr/0003-keycloak-as-idp.md), [adr/0004-postgres-goose-sqlc.md](../adr/0004-postgres-goose-sqlc.md)

## 1. Problem
Setor.in butuh sistem auth yang aman, mendukung 3 role (user, collector, admin),
dengan email verification & password reset, tanpa membangun sendiri (security
risk + bandwidth tim kecil).

## 2. Goals
| # | Goal | Acceptance |
|---|---|---|
| G1 | User register/login/reset password lewat Keycloak | Manual flow di MailHog dev |
| G2 | Backend Go validate JWT pakai JWKS, cache 1 jam | Latency validation <5ms p99 |
| G3 | RBAC dengan composite roles (user ⊂ collector ⊂ admin) | 403 untuk role insufficient |
| G4 | Local DB simpan profil; Keycloak simpan credential | Tidak ada kolom `password*` di DB Setor.in |
| G5 | Audit trail untuk semua event auth | Login/logout/role-change ter-log di `auth_events` |
| G6 | Collector role butuh approval admin | Endpoint `become-collector` create request, bukan langsung set role |

## 3. Non-Goals (v1)
- ❌ Social login (Google, Apple) — phase 2
- ❌ 2FA / TOTP — phase 2
- ❌ Magic link login — phase 2
- ❌ SSO ke aplikasi pihak ketiga — out of scope total
- ❌ Custom Keycloak theme/branding — phase 2 (pakai default)
- ❌ Multi-realm (per-city) — out of scope
- ❌ Self-service email change — phase 2
- ❌ Account deletion (GDPR-style) — phase 2 (untuk MVP, manual via admin)

## 4. User Stories

### Sebagai Aria (calon user baru)
- US-01: Saya bisa register dengan email + password + nama lengkap
- US-02: Saya menerima email verifikasi setelah register
- US-03: Setelah verify email, saya bisa login dan masuk ke aplikasi
- US-04: Saya tidak bisa login sebelum verify email (jelas pesan errornya)
- US-05: Kalau lupa password, saya bisa reset via email link

### Sebagai user existing
- US-06: Saya bisa login dengan email + password
- US-07: Saya bisa lihat profil saya (`GET /v1/auth/me`)
- US-08: Saya bisa logout (token invalid)
- US-09: Saya bisa request menjadi collector (upload dokumen + tunggu approval)

### Sebagai Budi (calon collector)
- US-10: Setelah register sebagai user, saya submit application untuk collector
- US-11: Saya tahu status application saya (pending/approved/rejected)
- US-12: Setelah approved, role saya otomatis upgrade — login berikutnya
  token saya berisi role `collector`

### Sebagai Admin
- US-13: Saya bisa lihat list semua user dengan filter & pagination
- US-14: Saya bisa lihat detail user (profile + role history)
- US-15: Saya bisa approve/reject collector application
- US-16: Saya bisa suspend/unsuspend user
- US-17: Saya bisa promote user lain jadi admin (super_admin only)

### Sistem (cross-cutting)
- US-18: Setiap event auth (login, logout, role change) ter-log di `auth_events`
- US-19: Token expired → backend return 401 dengan kode `TOKEN_EXPIRED`
- US-20: Role tidak cukup → backend return 403 dengan kode `INSUFFICIENT_ROLE`

## 5. Success Criteria
- 100% endpoint protected mengembalikan 401 tanpa token, 403 tanpa role
- Email verification mendarat di MailHog (dev) dalam <5 detik
- Sync Keycloak → local DB <500ms (p95)
- 0 password ter-simpan di DB Setor.in (verifikasi via grep `password` di
  schema migration)
- Test coverage usecase >80%

## 6. Roles & Permissions Matrix

| Endpoint | guest | user | collector | admin | super_admin |
|---|---|---|---|---|---|
| `POST /v1/auth/sync` | ✅ (with valid Keycloak token) | ✅ | ✅ | ✅ | ✅ |
| `GET /v1/auth/me` | ❌ | ✅ | ✅ | ✅ | ✅ |
| `POST /v1/auth/logout` | ❌ | ✅ | ✅ | ✅ | ✅ |
| `POST /v1/users/me/become-collector` | ❌ | ✅ | ❌ | ❌ | ❌ |
| `GET /v1/users/me` | ❌ | ✅ | ✅ | ✅ | ✅ |
| `PATCH /v1/users/me` | ❌ | ✅ | ✅ | ✅ | ✅ |
| `GET /v1/admin/users` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `GET /v1/admin/users/{id}` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `PATCH /v1/admin/users/{id}/status` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `PATCH /v1/admin/users/{id}/roles` | ❌ | ❌ | ❌ | ❌ | ✅ |
| `GET /v1/admin/collector-applications` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `POST /v1/admin/collector-applications/{id}/approve` | ❌ | ❌ | ❌ | ✅ | ✅ |
| `POST /v1/admin/collector-applications/{id}/reject` | ❌ | ❌ | ❌ | ✅ | ✅ |

Composite roles:
- `collector` includes `user`
- `admin` includes `user`
- `super_admin` includes `admin` (transitively `user`)

## 7. Open Questions

- [x] Email verification wajib sebelum login? **YES** (default Keycloak)
- [x] Collector approval butuh admin? **YES** (anti-fraud)
- [ ] Apakah user pertama (bootstrap) jadi `super_admin` otomatis? Atau lewat
      script seed? **Decision: lewat script seed** (`scripts/seed-superadmin.sh`)
- [ ] Berapa lama JWT lifetime? **Decision: access 15min, refresh 7day**
- [ ] Token frontend disimpan dimana? **Decision: lihat security.md** —
      tentatif httpOnly cookie + same-site lax untuk web

## 8. Out of Scope Detail

### Tidak akan dibangun di sprint ini
- UI registration/login (frontend Next.js — sprint berbeda)
- Email template custom (pakai default Keycloak)
- Notification push setelah role change (FCM — modul `notification`, sprint lain)

### Tidak akan dibangun pernah (architectural)
- Backend tidak pernah handle password
- Backend tidak pernah hash/verify password sendiri
- Backend tidak pernah issue JWT sendiri (Keycloak yang issue)
