# Database Migrations Log

Append-only changelog untuk setiap migration. Tujuan: audit, debugging, dan
backup planning.

Format: `<filename> — <YYYY-MM-DD> — <PR#> — <summary>`

---

## Sprint 0002 — Auth & User Management

| File | Date | PR | Summary |
|---|---|---|---|
| `00001_create_users.sql` | 2026-05-05 | TBD | Tabel `users` (link ke Keycloak), enum `user_role` & `user_status`, tabel `user_role_history`, trigger updated_at |
| `00002_create_auth_audit.sql` | 2026-05-05 | TBD | Tabel `auth_events` untuk audit trail event auth |
| `00003_create_collector_applications.sql` | 2026-05-05 | TBD | Tabel `collector_applications` + extension PostGIS, partial unique index untuk 1 pending application per user |

---

## Sprint 0003 — Catalog & Pricing

| File | Date | PR | Summary |
|---|---|---|---|
| `00004_create_materials.sql` | 2026-05-11 | TBD | Tabel `materials` (master 8 kategori) + `material_prices` (time-series, append-only). LATERAL join pattern untuk lookup current price |
| `00005_seed_materials.sql` | 2026-05-11 | TBD | Seed 8 initial materials (plastik, kardus, kertas, aluminium, tembaga, besi, kaca, e-waste) dengan harga awal dari design bundle |

---

## Sprint 0004 — Order Management (Cash)

| File | Date | PR | Summary |
|---|---|---|---|
| `00006_create_orders.sql` | 2026-05-11 | TBD | Tabel `orders` dengan 10-state enum + `order_status_history` (audit trail) + `order_photos`. Sequence `order_code_seq` untuk ECC-XXXXX. Kolom `payment_method` siap untuk wallet (Phase 2) tapi default 'cash' untuk MVP. |

| `00007_add_cv_role.sql` | 2026-05-18 | TBD | Tambah enum `cv` pada `user_role` agar role CV selaras dengan Keycloak dan middleware RBAC |

Catatan: migration `00006_create_wallets.sql` sempat dibuat untuk wallet module, lalu **rolled back & dihapus** karena pivot ke cash payment. Spec asli di-archive ke `specs/_archived/0004-wallet/`.

---

## Operational Notes

### Cara jalankan
```bash
make migrate         # apply semua pending
make migrate-down    # rollback 1 step
make migrate-status  # lihat status
make migrate-reset   # rollback semua (DEV ONLY)
```

### Kalau migration gagal di tengah
1. Cek log goose
2. Cek isi `goose_db_version` table
3. Manual fix kalau perlu, lalu update `goose_db_version` row terakhir
4. NEVER edit migration yang sudah di-apply di environment lain (tulis migration baru)

### Backup sebelum migration di prod
```bash
pg_dump -Fc -h $DB_HOST -U $DB_USER -d $DB_NAME -f backup-$(date +%Y%m%d-%H%M%S).dump
```

### Schema diff vs production
```bash
# Local vs prod
pg_dump --schema-only --no-owner postgres://prod-url > /tmp/prod.sql
pg_dump --schema-only --no-owner postgres://local-url > /tmp/local.sql
diff /tmp/prod.sql /tmp/local.sql
```
