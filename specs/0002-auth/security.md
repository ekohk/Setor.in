# Spec 0002 — Security Threat Model

## Scope
Auth & user management (sprint 0002). Threat to: backend Go API, Keycloak,
Postgres.

## STRIDE summary

| Threat | Mitigation |
|---|---|
| **S**poofing identity | JWT signed RS256, JWKS verification, issuer & audience check |
| **T**ampering | TLS in transit (prod), `iat`/`exp`/`nbf` validation |
| **R**epudiation | `auth_events` audit trail (append-only) |
| **I**nformation disclosure | Soft delete + `deleted_at` filter, no password in logs/DB |
| **D**enial of Service | Rate limit `/v1/auth/sync` (10 req/min/IP), Keycloak brute force lockout |
| **E**levation of privilege | RBAC middleware, `super_admin` only for role mutations |

## Specific threats & controls

### T1. Stolen access token
**Risk:** attacker replays JWT until expiry (15min).
**Mitigations:**
- Short-lived access token (15min)
- Refresh token rotation (single use, lifetime 7d)
- Frontend store: lihat T6 dibawah
- HSTS + TLS only di production
- Mark user-related endpoints with `Cache-Control: no-store`
- Idle session timeout di Keycloak (30min)

### T2. Stolen refresh token
**Risk:** attacker bisa generate access token baru terus-menerus.
**Mitigations:**
- Refresh token rotation (Keycloak default)
- Detect reuse (Keycloak invalidate session kalau reuse terdeteksi)
- Short refresh token lifetime (7 days)
- Refresh token disimpan ONLY di httpOnly cookie (web) / secure storage (mobile)

### T3. JWKS rotation gap
**Risk:** Keycloak rotate keys, backend cache stale → tolak token valid.
**Mitigations:**
- JWKS cache TTL 1 jam
- On unknown `kid`, force-refresh JWKS sekali, baru fail
- Alert kalau JWKS endpoint unreachable

### T4. Token from different realm/issuer
**Risk:** attacker pakai token dari realm lain yang validly signed.
**Mitigations:**
- WAJIB validate `iss == KEYCLOAK_ISSUER`
- WAJIB validate `aud` contains expected client ID
- WAJIB validate `azp` (authorized party)

### T5. Role tampering
**Risk:** client kirim claim role tambahan sendiri.
**Mitigations:**
- Backend HANYA percaya `realm_access.roles` di JWT yang **signature-verified**
- Tidak ada endpoint yang baca role dari header/body request
- Composite role expansion dilakukan di backend (cek `hasRole(roles, "admin")`
  yang juga return true untuk `super_admin`)

### T6. Token storage di frontend (XSS)
**Risk:** XSS curi token dari localStorage.
**Mitigations:**
- **Web:** pakai httpOnly cookie via BFF pattern
  - Frontend Next.js endpoint `/api/auth/callback` simpan token di httpOnly,
    secure, sameSite=lax cookie
  - Frontend ambil cookie pakai `credentials: 'include'`
- **Mobile (RN, phase 2):** Keychain (iOS) / EncryptedSharedPreferences (Android)
- **CSP header:** strict, no inline-script (Next.js sudah default)
- **NEVER** simpan token di localStorage atau sessionStorage

### T7. Email enumeration via forgot password
**Risk:** attacker enumerate email valid lewat error message.
**Mitigations:**
- Keycloak default: forgot password ALWAYS return success
- Backend `/v1/auth/sync`: jangan return error spesifik "user not found",
  cukup 401 generic

### T8. Mass user enumeration via admin endpoint
**Risk:** non-admin probe `/v1/admin/users`.
**Mitigations:**
- RBAC middleware (admin only)
- Rate limit per user
- Audit log every admin query

### T9. Privilege escalation via collector application
**Risk:** user submit application palsu untuk dapat role collector.
**Mitigations:**
- Manual admin approval (sudah di flow)
- Upload KTP + SIUP + foto tempat usaha (manual verify)
- Audit log siapa approve/reject + reason
- Rate limit: 1 application aktif per user (DB partial unique index)

### T10. Privilege escalation via admin promotion
**Risk:** admin biasa bisa promote user lain jadi admin.
**Mitigations:**
- Endpoint `/v1/admin/users/{id}/roles` HANYA `super_admin`
- Audit log + notification ke semua super_admin tiap promotion
- Bootstrap super_admin via seed script (manual), tidak ada self-service

### T11. SQL Injection
**Risk:** crafted input merusak query.
**Mitigations:**
- sqlc generate parameterized queries (compile-time guarantee)
- Tidak ada string concat untuk SQL
- Lint rule: ban `fmt.Sprintf` di file repo

### T12. Sensitive data in logs
**Risk:** PII (email, phone) atau token bocor di log.
**Mitigations:**
- Logger middleware redact `Authorization` header
- Redact `password` field (defense in depth — kita ga punya, tapi safety)
- Email log only domain (`****@gmail.com`)
- Structured logging (slog) dengan field allowlist

### T13. CSRF
**Risk:** kalau pakai cookie auth, attacker bisa CSRF.
**Mitigations:**
- SameSite=lax (default), atau strict untuk endpoint mutating
- Double-submit cookie pattern untuk POST/PATCH/DELETE
- Origin/Referer check di middleware
- Atau pakai Authorization header (bukan cookie) → CSRF tidak applicable

### T14. CORS misconfiguration
**Risk:** wildcard origin → attacker site bisa call API user.
**Mitigations:**
- Whitelist origin spesifik per env
- Tidak boleh `Access-Control-Allow-Origin: *` kalau credentials
- `OPTIONS` preflight di-handle benar

### T15. PII storage
**Risk:** UU PDP Indonesia.
**Mitigations:**
- Email & phone di-encrypt at rest? (target: Postgres TDE / KMS-encrypted disk)
- Audit log retention 1-2 tahun
- Right to be forgotten: hard delete via admin endpoint (phase 2)
- Data residency: server di Indonesia (AWS Jakarta region)

---

## Implementation requirements (acceptance)

Di sprint ini, WAJIB:

- [ ] Middleware verify JWT lewat JWKS (tidak introspect)
- [ ] Middleware reject token tanpa: signature valid, exp masih, iss benar, aud benar
- [ ] RBAC middleware: chain-able (`Require("admin")`), composite-aware
- [ ] Logger redact `Authorization` header
- [ ] No password column di schema (verifikasi: `grep -r 'password' db/migrations/`)
- [ ] No `fmt.Sprintf` for SQL
- [ ] Soft delete: semua repository query filter `WHERE deleted_at IS NULL`
- [ ] CORS allowlist via env var (no wildcard)
- [ ] Rate limit `/v1/auth/sync` (10/min/IP) — pakai middleware sederhana
- [ ] Audit log: register, login, logout, role_change, application_*

Di sprint berikutnya (di luar scope):

- [ ] CSP header
- [ ] HSTS
- [ ] httpOnly cookie BFF di frontend
- [ ] PII encryption at rest (DB level)
- [ ] WAF di production
