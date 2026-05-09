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
