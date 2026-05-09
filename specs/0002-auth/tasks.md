# Spec 0002 — Tasks Breakdown

Eksekusi berurutan. Tiap task = 1 PR (kecuali T1 & T2 boleh digabung).

| ID | Title | Acceptance | Estimasi |
|---|---|---|---|
| **T1** | docker-compose: keycloak + postgres-app + postgres-kc + mailhog | `make up` jalan, semua container healthy | 2h |
| **T2** | Goose migrations 00001 + 00002 + 00003 | `make migrate` clean, `make migrate-down` reverse clean | 1h |
| **T3** | Konfigurasi Keycloak realm setorin (manual via UI, lalu export) | `realm-export.json` tersimpan di repo, ulang import jalan | 2h |
| **T4** | `internal/shared/config` — viper config loader | Test load `.env` & validate required fields | 1h |
| **T5** | `internal/shared/database` — pgxpool init + health check | Test: connect, ping, close | 1h |
| **T6** | `internal/shared/logger` — slog setup, redact middleware | Log JSON, `Authorization` header tersamarkan | 1h |
| **T7** | `internal/shared/response` — envelope helper (Ok, Err, Paginated) | Unit test format konsisten | 1h |
| **T8** | `internal/infrastructure/keycloak/jwks.go` — JWKS validator dengan cache | Unit test: valid token pass, expired fail, wrong iss fail | 3h |
| **T9** | `internal/infrastructure/keycloak/admin_client.go` — gocloak wrapper | Unit/integration: get user, set role, revoke session | 3h |
| **T10** | `internal/shared/middleware/jwt.go` + `rbac.go` | Test: 401 no token, 401 invalid, 403 wrong role, 200 ok | 2h |
| **T11** | `internal/apps/user/domain/model/user.go` + enums + errors | Compile clean, lint pass | 1h |
| **T12** | `internal/apps/user/application/ports/user_repository.go` (interface) | Interface lengkap untuk usecase | 1h |
| **T13** | `internal/apps/user/adapters/persistence/postgres_user_repo.go` (impl) | Integration test pakai testcontainers/postgres | 3h |
| **T14** | `internal/apps/user/application/dto/*.go` — request/response | Validate tag lengkap | 1h |
| **T15** | `internal/apps/user/application/usecase/user_usecase.go` | Test: SyncFromKeycloak, GetMe, UpdateProfile | 3h |
| **T16** | `internal/apps/user/delivery/http/user_handler.go` + `routes.go` | curl test endpoint /v1/users/me | 2h |
| **T17** | `internal/apps/auth/delivery/http/auth_handler.go` + routes | Test /v1/auth/sync, /v1/auth/me, /v1/auth/logout | 3h |
| **T18** | Domain model `collector_application` + repository + usecase + handler | End-to-end POST → list → approve flow jalan | 4h |
| **T19** | `internal/apps/admin/...` — admin endpoints (list users, update status, update roles) | RBAC enforced, audit log tertulis | 4h |
| **T20** | `cmd/api/main.go` — wire semua dependency, bootstrap server | `go run` start tanpa error, http://localhost:8000/health 200 | 2h |
| **T21** | Postman collection untuk smoke test | Collection committed `docs/postman/auth.json` | 1h |
| **T22** | README quickstart 5-menit | Junior bisa setup dari nol dalam <30 menit | 1h |
| **T23** | Integration test (testcontainers) — full flow register→sync→me→admin | `go test ./...` pass di CI | 4h |

**Total estimasi: ~46 jam (~6 hari kerja).**

## Definition of Done per Task
- [ ] Code lulus `go build ./...`
- [ ] Code lulus `go vet ./...`
- [ ] Code lulus `golangci-lint run`
- [ ] Test coverage usecase ≥ 80%
- [ ] Tidak ada `fmt.Println` (pakai `slog`)
- [ ] Tidak ada hardcoded credential (semua via env)
- [ ] PR description sebutkan task ID + acceptance check
- [ ] `MIGRATIONS.md` updated kalau task touch DB
- [ ] OpenAPI `api.yaml` updated kalau task touch API

## Sprint cut points (kalau waktu mepet)

**Minimum viable:** T1–T17 (auth & user dasar jalan)
**MVP-ready:** T1–T22
**Production-ready:** T1–T23 + ADR baru untuk deployment, monitoring, secrets

## Dependencies graph

```
T1 (docker) → T2 (migrate) → T5 (db pool)
T1 → T3 (keycloak realm) → T8 (JWKS) → T10 (middleware)
T4 (config) → T5, T8, T9
T6 (logger) → T8, T10, T15, T17
T7 (response) → T16, T17, T19
T11 (model) → T12 (port) → T13 (impl)
T11 → T14 (dto) → T15 (usecase)
T15 → T16 (user handler) → T17 (auth handler)
T18, T19 → T17 (depends on auth middleware)
T20 wires everything
T21, T22, T23 → setelah T20
```
