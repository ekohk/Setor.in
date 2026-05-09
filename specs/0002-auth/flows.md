# Spec 0002 — Auth Flows (Sequence Diagrams)

Semua diagram pakai Mermaid. Render di Github / VS Code Mermaid preview.

---

## Flow 1 — Register

```mermaid
sequenceDiagram
  autonumber
  participant U as User (Browser)
  participant FE as Frontend (Next.js + keycloak-js)
  participant KC as Keycloak
  participant SMTP as Email (MailHog/Resend)
  participant BE as Backend Go

  U->>FE: Klik "Register"
  FE->>KC: Redirect ke /realms/setorin/protocol/openid-connect/registrations
  U->>KC: Isi form (email, password, full_name)
  KC->>KC: Create user (status: needs VERIFY_EMAIL)
  KC->>SMTP: Send verification email
  SMTP-->>U: Email dengan link verify
  KC-->>FE: Redirect dengan ?error=email_not_verified (atau success)
  U->>U: Klik link verifikasi di email
  U->>KC: GET /verify-email?key=...
  KC->>KC: Mark email verified
  KC-->>U: Halaman success → tombol Login

  Note over U,BE: Saat user pertama kali login (lihat Flow 2),<br/>backend create row di tabel `users` lokal
```

**Edge cases:**
- Email sudah terdaftar → Keycloak return error, frontend show "email already exists"
- Password tidak memenuhi policy → Keycloak return validation error
- Email gagal terkirim (SMTP down) → user belum bisa login, ada tombol "resend email" di Keycloak

---

## Flow 2 — Login (first time after verify)

```mermaid
sequenceDiagram
  autonumber
  participant U as User
  participant FE as Frontend
  participant KC as Keycloak
  participant BE as Backend Go
  participant DB as Postgres

  U->>FE: Submit login (email, password)
  FE->>KC: POST /realms/setorin/protocol/openid-connect/token<br/>(grant_type=password OR redirect-based PKCE)
  KC->>KC: Validate credential + email_verified
  KC-->>FE: { access_token, refresh_token, id_token }
  FE->>BE: POST /v1/auth/sync<br/>Authorization: Bearer <access_token>
  BE->>BE: Validate JWT (JWKS cached)
  BE->>DB: SELECT * FROM users WHERE keycloak_id = sub
  alt User belum ada di DB lokal
    BE->>KC: GET /admin/users/{sub} (via service account)
    KC-->>BE: User detail (email, given_name, family_name)
    BE->>DB: INSERT INTO users (keycloak_id, email, full_name, ...)
    BE->>DB: INSERT INTO auth_events (event_type='register_synced', ...)
  end
  BE->>DB: UPDATE users SET last_login_at=NOW()
  BE->>DB: INSERT INTO auth_events (event_type='login', ...)
  BE-->>FE: 200 { data: { user: {...}, roles: [...] } }
  FE->>FE: Store tokens (httpOnly cookie via BFF, lihat security.md)
  FE-->>U: Redirect ke Home
```

**Edge cases:**
- Email belum verified → Keycloak return 401 dengan `error=invalid_grant` & `error_description` mengandung "verify"
- User suspended di Keycloak → 401
- Token valid tapi user di-soft-delete di DB lokal (`deleted_at IS NOT NULL`) → `/v1/auth/sync` return 403 `ACCOUNT_DEACTIVATED`

---

## Flow 3 — Authenticated request (subsequent)

```mermaid
sequenceDiagram
  autonumber
  participant FE as Frontend
  participant BE as Backend Go
  participant KC as Keycloak (JWKS only)
  participant DB as Postgres

  FE->>BE: GET /v1/auth/me<br/>Authorization: Bearer <access_token>
  BE->>BE: JWT middleware: verify signature
  alt JWKS cache miss
    BE->>KC: GET /realms/setorin/protocol/openid-connect/certs
    KC-->>BE: JWKS
    BE->>BE: Cache 1 jam
  end
  BE->>BE: Verify exp, iss, aud
  BE->>BE: Extract sub, roles, db_user_id
  BE->>BE: RBAC middleware (check role)
  BE->>DB: SELECT user, profile WHERE keycloak_id=sub AND deleted_at IS NULL
  BE-->>FE: 200 { data: {...} }
```

---

## Flow 4 — Token refresh

```mermaid
sequenceDiagram
  autonumber
  participant FE as Frontend
  participant KC as Keycloak

  FE->>FE: Detect access_token expired (15 min)
  FE->>KC: POST /realms/setorin/protocol/openid-connect/token<br/>grant_type=refresh_token<br/>refresh_token=<rt>
  KC->>KC: Validate refresh token (rotation enabled)
  KC-->>FE: { new_access_token, new_refresh_token }
  Note over FE: Frontend rotate refresh_token (old one invalid)
```

**Edge cases:**
- Refresh token expired (>7 hari) → user harus login ulang
- Refresh token reused (replay) → Keycloak invalidate seluruh session

---

## Flow 5 — Forgot password

```mermaid
sequenceDiagram
  autonumber
  participant U as User
  participant FE as Frontend
  participant KC as Keycloak
  participant SMTP as Email

  U->>FE: Klik "Lupa password"
  FE->>KC: Redirect ke /realms/setorin/login-actions/reset-credentials
  U->>KC: Submit email
  KC->>SMTP: Send reset password email (kalau email terdaftar)
  Note over KC: Selalu return success generic (anti email-enumeration)
  KC-->>U: "Cek email Anda"
  SMTP-->>U: Email dengan link reset (token expire 5 menit)
  U->>U: Klik link
  U->>KC: GET /reset-credentials?key=...
  KC-->>U: Form password baru
  U->>KC: Submit password baru
  KC->>KC: Update credential, invalidate semua session existing
  KC-->>U: Redirect ke login
```

**Edge cases:**
- Email tidak terdaftar → tetap return success (anti enumeration), tapi tidak ada email terkirim
- Link expired → tampilkan "link kadaluarsa, request ulang"

---

## Flow 6 — Become collector (request)

```mermaid
sequenceDiagram
  autonumber
  participant U as User (logged in)
  participant FE as Frontend
  participant BE as Backend Go
  participant DB as Postgres
  participant ADMIN as Admin

  U->>FE: Buka "Switch to Collector mode"
  FE->>FE: Show form: business_name, license_no, KTP photo, SIUP photo
  U->>FE: Submit
  FE->>BE: POST /v1/users/me/become-collector<br/>(multipart with files)
  BE->>BE: RBAC check: must have role 'user', must NOT already 'collector'
  BE->>BE: Upload files to S3/MinIO
  BE->>DB: INSERT INTO collector_applications (user_id, status='pending', ...)
  BE->>DB: INSERT INTO auth_events (event_type='collector_application_submitted', ...)
  BE-->>FE: 201 { application_id, status: 'pending' }
  FE-->>U: "Application submitted. Tunggu approval admin."

  Note over BE,ADMIN: Async — admin review di dashboard

  ADMIN->>BE: POST /v1/admin/collector-applications/{id}/approve
  BE->>BE: RBAC check: admin or super_admin
  BE->>DB: UPDATE collector_applications SET status='approved'
  BE->>KC: PUT /admin/users/{user_id}/role-mappings/realm<br/>(add 'collector' role)
  BE->>DB: INSERT INTO user_role_history (from='user', to='collector', ...)
  BE->>DB: UPDATE users SET primary_role='collector'
  BE->>DB: INSERT INTO auth_events (event_type='role_changed', ...)
  BE-->>ADMIN: 200 OK
  Note over U: User akan dapat role baru di token saat refresh berikutnya
```

**Edge cases:**
- User sudah punya role collector → 409 Conflict
- Application pending sudah ada → 409 Conflict (1 application aktif per user)
- Reject → status `rejected`, user bisa apply ulang setelah X hari

---

## Flow 7 — Admin promote user → admin (super_admin only)

```mermaid
sequenceDiagram
  autonumber
  participant SA as Super Admin
  participant BE as Backend Go
  participant KC as Keycloak
  participant DB as Postgres

  SA->>BE: PATCH /v1/admin/users/{id}/roles<br/>{ "add": ["admin"] }
  BE->>BE: RBAC check: super_admin only
  BE->>BE: Validate target user exists & not super_admin
  BE->>KC: PUT /admin/users/{id}/role-mappings/realm
  BE->>DB: INSERT INTO user_role_history
  BE->>DB: UPDATE users SET primary_role='admin'
  BE->>DB: INSERT INTO auth_events (event_type='role_promoted', ...)
  BE-->>SA: 200 { user, new_roles }
```

---

## Flow 8 — Logout

```mermaid
sequenceDiagram
  autonumber
  participant FE as Frontend
  participant BE as Backend Go
  participant KC as Keycloak

  FE->>BE: POST /v1/auth/logout (with refresh_token in body)
  BE->>KC: POST /realms/setorin/protocol/openid-connect/logout<br/>(revoke refresh_token via service account)
  KC-->>BE: 204
  BE->>DB: INSERT INTO auth_events (event_type='logout', ...)
  BE-->>FE: 204
  FE->>FE: Clear local tokens
```

---

## Notes

- **Semua flow yang melibatkan email** → di dev MailHog (port 8025 UI, 1025 SMTP)
- **JWKS cache TTL** → 1 jam, tapi dengan force refresh kalau ketemu `kid`
  unknown (handle key rotation)
- **Clock skew tolerance** → 60 detik untuk `exp` dan `nbf`
- **Audit log** dipanggil di tiap flow yang signifikan — implementasi via
  decorator/middleware di usecase layer
