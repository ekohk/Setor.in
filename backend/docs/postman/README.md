# Postman Collection — Setor.in API

> **Aturan**: 1 collection saja. Tiap fitur baru, file ini di-update — Anda
> tinggal re-import. Hanya happy path (status 200/201/204) — error case
> di-cover di unit test, bukan Postman.

## ✨ Auto-Token Magic

Collection ini punya **pre-request script di level collection** yang otomatis
ambil token baru dari Keycloak SETIAP request kalau token di env kosong atau
sebentar lagi expired (30 detik buffer).

Artinya:
- **Anda tidak perlu jalankan Login dulu** — langsung klik request apapun
- **Token expired? Auto-refresh** sebelum request dieksekusi
- **Ganti user (admin/super_admin)?** Edit `user_email` + `user_password` di
  env, lalu klik request → token baru otomatis fetched untuk user tsb

Folder `01 — Auth (Keycloak)` masih ada untuk **manual login** (kalau Anda mau
inspect raw Keycloak response), tapi jarang dipakai sehari-hari.

## File di folder ini

| File | Isi |
|---|---|
| `Setor.in.postman_collection.json` | 1 collection berisi semua endpoint |
| `Setor.in.postman_environment.json` | Environment variables |
| `README.md` | Panduan ini |

---

# Setup (sekali saja)

## 1. Import 2 file

Buka Postman → **File → Import** → drag dua JSON ini.

## 2. Pilih environment

Top-right dropdown → **Setor.in Local Dev**.

## 3. Pastikan backend & docker jalan

Di terminal:
```powershell
cd c:\Users\HP\OneDrive\Documents\Setor.in\backend
make up                # docker (Postgres + Keycloak + MailHog)
go run ./cmd/api       # backend (terminal lain, biarkan jalan)
```

## 4. Pastikan user test ada di Keycloak

Default collection pakai `eko@test.local` / `Eko1234!`.

Cara cek/buat:
1. Buka http://localhost:8090/admin
2. Login `admin` / `admin`
3. **Top-left dropdown** → pastikan realm = **Setor.in** (BUKAN "Keycloak")
4. Sidebar **Users** → cari `eko@test.local`. Kalau tidak ada:
   - Klik **Add user**
   - Email: `eko@test.local`
   - Username: `eko@test.local`
   - **Email verified**: ON
   - First name: `eko`, Last name: `hadi`
   - Klik **Create**
   - Tab **Credentials** → **Set password**:
     - Password: `Eko1234!`
     - **Temporary**: OFF (penting!)
     - Save → konfirmasi

---

# Environment Variables — Apa Isinya & Dari Mana?

Klik icon **mata** (eye) di pojok kanan-atas Postman untuk lihat semua variable.

## Variables yang Anda ISI MANUAL (sekali saja)

Sebagian besar sudah di-pre-fill dengan nilai default yang cocok untuk dev lokal Anda. **Tidak perlu diubah** kecuali Anda mau.

| Variable | Nilai default | Dari mana? Kapan ubah? |
|---|---|---|
| `api_base_url` | `http://localhost:8000` | URL backend Setor.in. Ubah kalau Anda jalankan backend di port lain (lihat `APP_PORT` di `.env`). |
| `kc_base_url` | `http://localhost:8090` | URL Keycloak. Ubah kalau Anda jalankan Keycloak di port lain (lihat `docker-compose.yml`). |
| `kc_realm` | `setorin` | Nama realm di Keycloak. **Jangan ubah** — sudah hardcoded saat setup realm di Sub-step 4.1. |
| `kc_web_client_id` | `setorin-web` | Client ID di Keycloak realm. **Jangan ubah** — sudah hardcoded saat buat client di Sub-step 4.7. |
| `user_email` | `eko@test.local` | Email user yang dipakai login. Ganti kalau Anda mau test pakai user lain. |
| `user_password` | `Eko1234!` | Password user. **Ganti kalau password user Anda berbeda**. |

## Variables yang AUTO-FILL (jangan diisi manual)

Ini di-set otomatis oleh test script saat Anda jalankan request tertentu. Kalau kosong, jalankan dulu request yang menghasilkannya.

| Variable | Diisi otomatis oleh | Untuk apa? |
|---|---|---|
| `access_token` | `01 → Login` | Bearer token untuk semua request authed |
| `refresh_token` | `01 → Login` | Untuk request `01 → Refresh token` |
| `keycloak_id` | `01 → Login` | UUID user di Keycloak (= JWT `sub`) |
| `user_id` | `02 → POST /v1/auth/sync` | UUID user di DB Setor.in (PK tabel `users`) |
| `application_id` | `03 → POST become-collector` ATAU `04 → GET collector-applications` | UUID collector application untuk approve/reject |

**Tips**: Kalau variable autofill kosong dan request gagal — cek folder mana yang harus dijalankan dulu (lihat tabel di atas).

---

# Cara Pakai — Workflow Standar

## Workflow 1: Test Sebagai User Biasa

1. **`01 → Login`** — dapat token, otomatis save
2. **`02 → POST /v1/auth/sync`** — pertama kali ini buat row di DB Setor.in
3. **`02 → GET /v1/auth/me`** — cek profil
4. **`02 → PATCH /v1/users/me`** — update full_name + phone
5. **`03 → POST become-collector`** — submit application

✅ Semua harus status 200/201.

## Workflow 2: Test Sebagai Admin

Kalau mau test endpoint folder `04 — Admin` dan `05 — Admin role mutation`, user Anda harus punya role `admin` (atau `super_admin`).

### A. Promote user di Keycloak (sekali saja)

1. http://localhost:8090/admin → admin/admin
2. **Top-left dropdown** → **Setor.in** (jangan stuck di "Keycloak"!)
3. Sidebar **Users** → klik user Anda (`eko@test.local`)
4. Tab **Role mapping** → klik **Assign role**
5. Filter: "Filter by realm roles" → centang **`admin`** (untuk folder 04) atau **`super_admin`** (untuk folder 05)
6. Klik **Assign**

### B. Re-login di Postman

⚠️ Token lama tidak akan otomatis dapat role baru. **Wajib login ulang**:

- **`01 → Login`** ulang

Buka **View → Show Postman Console** (Ctrl+Alt+C) — Anda harus lihat console log `Roles: ..., admin` atau `..., super_admin`.

### C. Jalankan endpoint admin

- **`04 → GET /v1/admin/users`** — list semua user
- **`04 → GET /v1/admin/collector-applications?status=pending`** — list pending application; otomatis save `application_id` ke env
- **`04 → POST .../approve`** — approve application yang dipilih di langkah sebelumnya. Setelah approve, role `collector` di-assign ke user di Keycloak.
- **`04 → PATCH .../status — suspend`** — suspend user (otomatis revoke session di Keycloak)
- **`04 → PATCH .../status — unsuspend`** — kembalikan ke active

### D. (Opsional) Test role mutation (super_admin)

Hanya kalau user Anda sudah punya `super_admin`:
- **`05 → PATCH /v1/admin/users/:id/roles`** — promote user ke `admin`

---

# Token Expired? (15 menit)

Access token cuma valid 15 menit. Kalau request dapat 401:

- Jalankan **`01 → Refresh token`** (lebih cepat dari login)
- Atau **`01 → Login`** ulang

---

# Lihat Decoded JWT (untuk debug)

1. **View → Show Postman Console** (Ctrl+Alt+C)
2. Setelah jalankan **`01 → Login`**, console log akan tampilkan email + roles
3. Atau decode manual: copy `access_token` → paste ke https://jwt.io

---

# Cara Tambah Endpoint Baru ke Collection (untuk maintainer)

Saat saya update collection (tiap fitur baru):
1. File `Setor.in.postman_collection.json` di-overwrite
2. Anda re-import: **File → Import** → pilih file → klik **Replace**
3. Postman ganti collection lama. Environment & token tetap.

Tidak perlu re-setup environment.

---

# Connect ke Database Tanpa psql

Pakai DBeaver / pgAdmin / TablePlus:

| Field | Value |
|---|---|
| Host | `localhost` |
| Port | `5435` |
| Database | `setorin` |
| Username | `setorin` |
| Password | `setorin_dev_pw` |
| SSL | disable |

Lihat tabel `users`, `auth_events`, `collector_applications` untuk verify hasil request Postman tersimpan di DB.
