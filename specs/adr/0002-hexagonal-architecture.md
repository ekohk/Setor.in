# ADR-0002: Hexagonal Architecture for Backend

## Status
Accepted — 2026-05-05

## Context
Backend Setor.in akan punya banyak modul (auth, user, order, wallet, collector,
payment, dll.) yang harus:
- Mudah di-test secara isolasi
- Mudah ditambah modul baru tanpa coupling
- Bisa swap implementation (mis. Postgres → Spanner, atau Xendit → Midtrans)
  tanpa rewrite use case

Tim sudah punya implementasi Hexagonal Architecture matang di **HRIS Soluix
recruitment module** (`c:/Documents/SOLUIX/HRIS-Soluix/hris-api/internal/apps/recruitment/`).

## Decision
Adopt **Hexagonal / Clean Architecture** dengan layout 5-folder per module,
**mengikuti pola HRIS recruitment**:

```
internal/apps/<module>/
├── domain/               ← Inti bisnis, no I/O, no framework
│   ├── model/            ← Entity (struct + behavior)
│   ├── services/         ← Domain services (logika murni cross-entity)
│   ├── events/           ← Domain events
│   └── errors.go
├── application/          ← Use cases (orchestrator)
│   ├── dto/              ← Request/response DTOs
│   ├── ports/            ← Interface (ke repo & external)
│   ├── usecase/          ← Business orchestration
│   └── services/         ← Application service (cross-aggregate)
├── adapters/             ← Implementasi ports
│   ├── persistence/      ← Postgres repo
│   └── external/         ← API client (Keycloak, Xendit, dst.)
├── delivery/             ← Input layer
│   └── http/             ← Gin handlers + routes.go
└── infrastructure/       ← Infra spesifik (jarang dipakai, biasanya kosong
                              kecuali butuh migration/script khusus modul)
```

## Aturan Dependensi (WAJIB)

```
delivery     → application       (boleh)
adapters     → application       (implements ports)
application  → domain            (boleh, satu arah)
infrastructure → adapters        (boleh)

domain       → application       ❌ DILARANG
domain       → adapters          ❌ DILARANG
domain       → delivery          ❌ DILARANG
application  → adapters          ❌ DILARANG (lewat ports!)
application  → delivery          ❌ DILARANG
```

Domain layer tidak boleh import:
- `gin`, `pgx`, `database/sql`
- Library HTTP, DB, queue apapun
- Hanya boleh import: stdlib + `uuid` + struct dari modul lain (model only)

Application/usecase tidak boleh import library DB langsung — harus lewat
interface di `application/ports/`.

## Consequences

### Positif
- Unit test usecase mudah (mock ports dengan testify/mock atau gomock)
- Bisa swap Postgres → SQLite untuk integration test
- Konsisten dengan HRIS — onboarding engineer baru cepat
- Domain bisa di-extract jadi shared library kalau perlu

### Negatif
- Boilerplate lebih banyak vs simple "handler-service-repo"
- Folder lebih dalam (5 layer)
- Junior dev perlu 1-2 minggu paham aturan dependensi

## Alternatives Considered

### A. Simple 3-layer (`handler/service/repository`)
- **Ditolak**: tidak skalabel saat modul banyak. Tidak ada batas jelas antara
  domain logic vs orchestration. Sulit test.

### B. DDD penuh dengan aggregate root, value object, dst.
- **Ditolak**: overkill untuk MVP. Bisa diadopsi gradual kalau model bisnis
  jadi kompleks (mis. order dengan banyak entity).

### C. Microservices dari hari pertama
- **Ditolak**: premature. Tim 1-2 orang, deployment overhead besar, latency
  inter-service merugikan UX.

## Module list (target Phase 1)
- `auth` — JWT validation, RBAC middleware (sekarang)
- `user` — profile, role management (sekarang)
- `catalog` — material, harga
- `order` — order lifecycle 8-stage
- `pickup` — collector matching, geo
- `weighing` — weight + quality
- `wallet` — ledger, balance
- `payment` — Xendit integration
- `collector` — onboarding, dashboard
- `notification` — FCM dispatcher
- `admin` — internal operations

## Reference
Lihat HRIS recruitment sebagai contoh living implementation:
`c:/Documents/SOLUIX/HRIS-Soluix/hris-api/internal/apps/recruitment/`
