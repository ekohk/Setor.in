# ADR-0004: Postgres + Goose + sqlc

## Status
Accepted — 2026-05-05

## Context
Backend Setor.in butuh:
- Database relational (transaksi finansial, FK constraints, ACID)
- Geo-search (collector terdekat) → Postgres + PostGIS
- Migration tool yang sistematis & auditable (goose)
- Type-safe DB access tanpa ORM overhead (sqlc)

## Decision

### 1. Database: **PostgreSQL 16 + PostGIS** extension
Versi 16 stable, performa baik, JSON type matang, PostGIS untuk geo-search
collector.

### 2. Driver: **`pgx/v5`**
- Lebih cepat dari `database/sql + lib/pq`
- Native Postgres type support (UUID, JSONB, INET, etc.)
- Connection pool built-in (`pgxpool`)

### 3. Migration: **goose** (SQL format only)
- Tracking di tabel `goose_db_version` di DB
- File naming: `00001_<snake_case>.sql`
- Format SQL (bukan Go) → portable, mudah di-review, bisa dijalankan manual
- Up & Down WAJIB di setiap migration (rollback safety)

### 4. Query layer: **sqlc**
- Generate Go code dari SQL → compile-time type safety
- No ORM magic (tidak ada N+1 hidden, tidak ada lazy loading)
- Hand-written SQL → query plan jelas
- Output ke `internal/apps/<module>/adapters/persistence/sqlc/` per modul

### 5. Migration vs Seed (PENTING)
- **Migration** = perubahan SCHEMA (DDL), versioned, irreversible-friendly
- **Seed** = data DUMMY untuk dev/test, reversible, di folder terpisah
- **Reference data** (master data: list material, list city, dll.) =
  migration TERPISAH dengan suffix `_seed_data` agar jelas

## Folder layout

```
backend/db/
├── migrations/
│   ├── 00001_create_users.sql
│   ├── 00002_create_auth_audit.sql
│   ├── 00003_create_materials.sql
│   ├── 00004_seed_materials.sql       ← reference data
│   └── MIGRATIONS.md                   ← changelog wajib
├── queries/
│   ├── users.sql                       ← input sqlc
│   └── auth_events.sql
├── seeds/
│   └── dev_users.sql                   ← dummy data, manual
└── sqlc.yaml
```

## Naming convention

### Migration
- Format: `<5-digit-sequence>_<verb>_<object>.sql`
- Verb: `create`, `add`, `drop`, `alter`, `rename`, `seed`
- Examples:
  - `00001_create_users.sql`
  - `00005_add_phone_to_users.sql`
  - `00012_drop_legacy_session_table.sql`
  - `00020_seed_initial_materials.sql`

### sqlc query
- File per table: `users.sql`, `orders.sql`
- Query name: `<Action><Object>` PascalCase
  - `GetUserByID`, `ListUsersByRole`, `CreateUser`, `UpdateUserStatus`
  - `:one`, `:many`, `:exec`, `:execrows`

## Migration policy (operational)

| Aturan | Kenapa |
|---|---|
| **Tiap migration WAJIB punya Down section** | Rollback safety di production |
| **Tidak boleh edit migration yang sudah deployed** | Idempotency, drift |
| **Breaking change wajib 2-step**: deploy compatible-add → migrate data → deploy remove | Zero downtime |
| **Foreign key wajib `ON DELETE` jelas** (RESTRICT/CASCADE/SET NULL) | Data integrity |
| **Index DDL pakai `CONCURRENTLY` di prod migration** | Tidak lock table besar |
| **Update `MIGRATIONS.md`** tiap PR yang tambah migration | Audit trail |

## Backup strategy (dev)
- `pg_dump` script di `backend/scripts/backup-dev-db.sh`
- Dump `goose_db_version` table → tahu seq terakhir
- Dump tiap migration ada hash (di MIGRATIONS.md)

## Backup strategy (prod, future)
- Daily logical backup ke S3
- Point-in-time recovery (WAL archive)
- Replica streaming (Postgres native)

## Consequences

### Positif
- Query jelas, performance jelas, no ORM magic
- Type-safe — compiler tangkap error sebelum runtime
- Migration auditable (file + git history + DB tracking)
- Mudah onboard engineer (SQL adalah lingua franca)

### Negatif
- Tulis SQL manual lebih lama vs `db.Find(&users)` ORM
- sqlc generate output perlu di-commit (tambah noise di PR)
- Refactor schema → harus update query SQL manual

## Alternatives Considered

### A. GORM
- **Ditolak**: hidden N+1, performance unpredictable, debugging sulit, magic
  hooks (BeforeSave, AfterFind) bikin susah test.

### B. Ent (Facebook)
- **Dipertimbangkan**. Schema-as-code, type-safe. **Ditolak** karena: learning
  curve, schema definition di Go (kita prefer SQL native), tim sudah familiar
  pola HRIS yang pakai `database/sql` murni.

### C. database/sql murni (seperti HRIS)
- **Dipertimbangkan kuat** karena konsistensi dengan HRIS. **Ditolak** karena:
  hand-write `rows.Scan()` repetitive & error-prone, sqlc menghilangkan boilerplate
  TANPA mengubah pola "SQL-first". sqlc tetap menghasilkan kode yang setara
  dengan database/sql murni — hanya generated.

### D. Migration dengan Atlas (HCL declarative)
- **Ditolak**: imperative migration (goose) lebih predictable untuk DB
  finansial. Atlas declarative bisa generate migration yang tidak diharapkan.

## Open Questions
- [ ] Connection pool size production: tentukan setelah load test
- [ ] Read replica: kapan dibutuhkan (estimasi >10k QPS)
- [ ] Partitioning strategy untuk `auth_events` & `orders` (estimasi 1 tahun)
