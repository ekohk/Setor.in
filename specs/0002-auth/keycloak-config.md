# Spec 0002 — Keycloak Configuration

Versi target: **Keycloak 25.x**

## 1. Realm

| Setting | Value |
|---|---|
| Realm name | `setorin` |
| Display name | `Setor.in` |
| Enabled | Yes |
| User-managed access | No (v1) |

### Login settings
- User registration: **enabled**
- Forgot password: **enabled**
- Remember me: enabled
- Email as username: **YES** (UX simpler)
- Login with email: YES
- Duplicate emails: NO
- Verify email: **REQUIRED** (cannot login without verifying)
- Edit username: NO (since username = email)

### Token settings
- Default Signature Algorithm: `RS256`
- SSO Session Idle: 30 minutes
- SSO Session Max: 7 days
- Access Token Lifespan: **15 minutes**
- Refresh Token Max Reuse: 0 (rotation enabled)
- Refresh Token Lifespan: 7 days

### Brute force detection (recommended ON for prod)
- Permanent lockout: NO
- Max login failures: 10
- Wait increment: 60s
- Quick login check: 1000ms
- Minimum quick login wait: 60s

### Password policy
```
length(8) and upperCase(1) and digits(1) and specialChars(1) and notUsername and passwordHistory(3)
```

---

## 2. Clients

### Client A: `setorin-web` (public — for frontend)

| Setting | Value |
|---|---|
| Client ID | `setorin-web` |
| Client type | OpenID Connect |
| Client authentication | **OFF** (public) |
| Authorization | OFF |
| Standard flow | ON (Authorization Code + PKCE) |
| Direct access grants | ON (untuk dev/testing dengan password grant) |
| Implicit flow | OFF |
| Service accounts | OFF |
| Valid redirect URIs | `http://localhost:3000/*`, `https://app.setor.in/*` |
| Web origins | `http://localhost:3000`, `https://app.setor.in` |
| Front channel logout | ON |
| PKCE | **REQUIRED** (`S256`) |

### Client B: `setorin-backend` (confidential — for backend admin operations)

| Setting | Value |
|---|---|
| Client ID | `setorin-backend` |
| Client type | OpenID Connect |
| Client authentication | **ON** (confidential) |
| Authorization | OFF (kita pakai realm roles, bukan resource server policies) |
| Standard flow | OFF |
| Direct access grants | OFF |
| Service accounts | **ON** |
| Valid redirect URIs | (kosong) |
| Service account roles | `manage-users`, `view-users`, `view-realm`, `query-users`, `query-groups` (dari `realm-management` client) |

Backend pakai client_credentials grant ke service account untuk:
- Lookup user detail by sub (saat /auth/sync)
- Add/remove realm roles (saat collector approval, role promotion)
- Revoke session (saat /auth/logout, saat suspend)

---

## 3. Realm Roles

Buat 4 role di Realm Roles (BUKAN client roles):

| Role | Composite | Includes |
|---|---|---|
| `user` | No | — |
| `collector` | **Yes** | `user` |
| `admin` | **Yes** | `user` |
| `super_admin` | **Yes** | `admin` (transitively `user`) |

### Default role
- Default Roles tab: `default-roles-setorin` includes `user`
  → setiap user yang register otomatis dapat role `user`

---

## 4. Required Actions

- `VERIFY_EMAIL` — enabled, default action
- `UPDATE_PASSWORD` — enabled (untuk reset)
- `terms_and_conditions` — disabled di MVP

---

## 5. Email / SMTP

### Dev (MailHog)
| Field | Value |
|---|---|
| Host | `mailhog` (kalau Keycloak di docker-compose yang sama) atau `host.docker.internal` |
| Port | `1025` |
| From | `noreply@setor.in` |
| From Display Name | `Setor.in` |
| Reply-to | `support@setor.in` |
| Enable SSL | NO |
| Enable StartTLS | NO |
| Authentication | NO |

### Prod (target Resend)
| Field | Value |
|---|---|
| Host | `smtp.resend.com` |
| Port | `587` |
| From | `noreply@setor.in` |
| Enable StartTLS | YES |
| Authentication | YES (username `resend`, password = API key) |

---

## 6. Custom Token Mapper (optional, recommended)

Tambah mapper di `setorin-web` client untuk include `db_user_id` di JWT,
agar backend bisa skip lookup `users` table untuk endpoint sederhana.

Approach (eksekusi setelah backend create row di DB):
1. Backend saat first sync → simpan `users.id` (UUID) sebagai user attribute
   `db_user_id` di Keycloak (via service account)
2. Tambah mapper di client scope:
   - Name: `db-user-id`
   - Mapper Type: `User Attribute`
   - User Attribute: `db_user_id`
   - Token Claim Name: `db_user_id`
   - Add to ID/Access token: YES

**Catatan:** kalau belum di-set, backend tetap fallback ke lookup `WHERE
keycloak_id = sub`. Mapper ini optimasi, bukan hard requirement.

---

## 7. Realm Export / Import

Untuk versioning config:

```bash
# Export (sekali setelah konfig manual)
docker exec -it setorin-keycloak \
  /opt/keycloak/bin/kc.sh export \
  --realm setorin \
  --file /tmp/realm-export.json

docker cp setorin-keycloak:/tmp/realm-export.json ./docker/keycloak/realm-export.json
```

Saat startup ulang, mount file ini ke Keycloak agar realm auto-import:

```yaml
# docker-compose.yml fragment
services:
  keycloak:
    command: start-dev --import-realm
    volumes:
      - ./keycloak/realm-export.json:/opt/keycloak/data/import/realm-export.json
```

---

## 8. Setup Checklist (manual via Admin UI, sekali di awal)

- [ ] Login http://localhost:8090 (admin/admin dari env)
- [ ] Create realm `setorin`
- [ ] Set login settings (verify email REQUIRED, dst.)
- [ ] Set token settings (15min/7day)
- [ ] Set password policy
- [ ] Create realm roles: user, collector, admin, super_admin
- [ ] Set composite: collector→user, admin→user, super_admin→admin
- [ ] Add `user` to default-roles-setorin
- [ ] Create client `setorin-web` (public, PKCE)
- [ ] Create client `setorin-backend` (confidential, service account)
- [ ] Assign service account roles ke setorin-backend (manage-users dst.)
- [ ] Configure SMTP (MailHog)
- [ ] Test register → email landing di MailHog UI :8025
- [ ] Test login → dapat token
- [ ] Export realm → commit `docker/keycloak/realm-export.json`

---

## 9. Env vars yang dibutuhkan backend

```bash
KEYCLOAK_BASE_URL=http://localhost:8090
KEYCLOAK_REALM=setorin
KEYCLOAK_BACKEND_CLIENT_ID=setorin-backend
KEYCLOAK_BACKEND_CLIENT_SECRET=<from-Keycloak-admin-UI>
KEYCLOAK_WEB_CLIENT_ID=setorin-web

# Derived
KEYCLOAK_ISSUER=http://localhost:8090/realms/setorin
KEYCLOAK_JWKS_URL=http://localhost:8090/realms/setorin/protocol/openid-connect/certs
KEYCLOAK_TOKEN_URL=http://localhost:8090/realms/setorin/protocol/openid-connect/token
```
