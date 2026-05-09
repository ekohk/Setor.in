# Setor.in Backend

Go modular monolith untuk Setor.in marketplace daur ulang.
Mengikuti **Hexagonal Architecture** seperti HRIS Soluix recruitment module.

## Quickstart (5 menit)

### 1. Prasyarat
- Go 1.23+
- Docker Desktop (running)
- Tools: `goose`, `sqlc`, `oapi-codegen`, `air`, `golangci-lint`

```bash
# Install tools sekali saja
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/air-verse/air@latest
```

### 2. Setup environment
```bash
cp .env.example .env
# Edit .env kalau perlu (default sudah OK untuk dev lokal)
```

### 3. Start infrastructure
```bash
make up
```
Tunggu ~60 detik sampai Keycloak healthy. Cek di `make ps`.

### 4. Konfigurasi Keycloak (sekali saja)
1. Buka http://localhost:8090 → login `admin` / `admin`
2. Ikuti checklist di [`../specs/0002-auth/keycloak-config.md`](../specs/0002-auth/keycloak-config.md)
3. Setelah setup selesai: `make keycloak-export` untuk versioning realm

### 5. Run migrations
```bash
make migrate
make migrate-status
```

### 6. Install Go deps & run
```bash
make tidy
make dev      # hot reload
# ATAU
make run      # build + run sekali
```

API akan jalan di http://localhost:8000

### 7. Test
```bash
curl http://localhost:8000/health
```

---

## Struktur Folder (Hexagonal Architecture)

```
backend/
├── cmd/
│   ├── api/              ← entry point HTTP server
│   └── migrate/          ← (opsional) wrapper goose
├── internal/
│   ├── apps/             ← SETIAP MODUL bisnis = 1 folder
│   │   ├── auth/         ← JWT validation, RBAC
│   │   │   ├── domain/{model,services,events}/
│   │   │   ├── application/{dto,ports,usecase}/
│   │   │   ├── adapters/{persistence,external}/
│   │   │   └── delivery/http/
│   │   └── user/         ← profile, role management
│   │       └── (struktur sama)
│   ├── shared/           ← cross-cutting concerns
│   │   ├── config/
│   │   ├── database/
│   │   ├── errors/
│   │   ├── logger/
│   │   ├── middleware/   ← JWT, RBAC, CORS, recover
│   │   ├── response/     ← envelope { data, meta, error }
│   │   └── validation/
│   ├── infrastructure/   ← integrasi dengan dunia luar
│   │   └── keycloak/     ← JWKS validator + admin client (gocloak)
│   └── services/         ← service lintas-modul (email, FCM)
├── pkg/                  ← reusable utilities
├── db/
│   ├── migrations/       ← goose SQL files
│   ├── queries/          ← sqlc input
│   └── sqlc.yaml
├── api/
│   └── openapi.yaml      ← single source of truth API
├── docker/
│   ├── docker-compose.yml
│   └── keycloak/
│       └── realm-export.json
└── Makefile
```

### Aturan dependensi (PENTING)
Lihat [`../specs/adr/0002-hexagonal-architecture.md`](../specs/adr/0002-hexagonal-architecture.md).
Singkatnya:

```
delivery → application → domain
adapters → application (implements ports) → domain
```

Domain TIDAK BOLEH import application/adapters/delivery.
Application TIDAK BOLEH import library DB/HTTP — hanya lewat `application/ports/`.

---

## Module list

| Module | Status | Spec |
|---|---|---|
| `auth` | In progress | [0002-auth/](../specs/0002-auth/) |
| `user` | In progress | [0002-auth/](../specs/0002-auth/) |
| `catalog` | Planned | TBD |
| `order` | Planned | TBD |
| `wallet` | Planned | TBD |
| `payment` | Planned | TBD |
| `notification` | Planned | TBD |

---

## Daily Commands

```bash
make up                           # Start docker services
make down                         # Stop docker services
make logs                         # Tail all logs
make keycloak-logs                # Tail Keycloak only

make migrate                      # Apply migrations
make migrate-down                 # Rollback 1
make migrate-status               # Show status
make migrate-create name=xxx      # Buat migration baru

make sqlc-gen                     # Regenerate Go code dari SQL queries
make tidy                         # go mod tidy
make build / run / dev            # Build / run / hot-reload
make test                         # go test -race ./...
make lint                         # golangci-lint
```

---

## Troubleshooting

### Keycloak takes forever to start
Normal — JVM startup ~30-60s. Cek `make keycloak-logs` sampai lihat
`Listening on: http://0.0.0.0:8090`.

### Port 5435/5434/8090/1025/8025 already in use
Edit `.env` & `docker/docker-compose.yml` ke port lain. Atau matikan service
yang pakai port tersebut.

### Goose error "no migrations found"
Pastikan `DB_DSN` di `.env` benar. Atau jalankan langsung:
```bash
goose -dir ./db/migrations postgres "postgres://setorin:setorin_dev_pw@localhost:5435/setorin?sslmode=disable" status
```

### Email tidak masuk MailHog
Cek SMTP setting di Keycloak realm: Realm Settings → Email.
Host harus `mailhog` (kalau Keycloak di docker-compose yang sama) atau
`host.docker.internal`.

---

## Specs

Semua keputusan arsitektur & desain ada di [`../specs/`](../specs/):

- [`0001-platform/spec.md`](../specs/0001-platform/spec.md) — Platform overview
- [`0002-auth/`](../specs/0002-auth/) — Auth & user management (sprint sekarang)
- [`adr/`](../specs/adr/) — Architecture Decision Records

---

## Reference Implementation

Pola hexagonal architecture mengikuti **HRIS Soluix recruitment module**:
`c:/Documents/SOLUIX/HRIS-Soluix/hris-api/internal/apps/recruitment/`
