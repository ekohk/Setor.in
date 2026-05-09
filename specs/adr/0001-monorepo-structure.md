# ADR-0001: Monorepo Structure

## Status
Accepted — 2026-05-05

## Context
Setor.in akan punya beberapa deliverable:
- Frontend web (Next.js, PWA mobile-first)
- Backend API (Go, modular monolith)
- Mobile app (React Native, phase 2)
- Admin panel (Next.js, phase 1.5)
- Shared types & API client (TypeScript)
- Design tokens dari handoff Claude Design

Pertanyaan: satu repo atau multiple repo?

## Decision
**Monorepo** dengan struktur:

```
Setor.in/
├── backend/                ← Go (sekarang)
├── apps/
│   ├── web/                ← Next.js (phase 1)
│   ├── mobile/             ← React Native (phase 2, kosong dulu)
│   └── admin/              ← Next.js admin (phase 1.5)
├── packages/
│   ├── api-client/         ← auto-generated dari OpenAPI
│   ├── types/              ← shared TS types
│   ├── ui/                 ← React components
│   └── design-tokens/      ← warna/spacing dari styles.css
├── specs/                  ← SDD documents (sekarang)
├── project/                ← design handoff (existing)
└── docs/
```

Backend Go tetap pakai layout idiomatic Go (`cmd/`, `internal/`, `pkg/`).
Frontend nanti pakai pnpm workspaces atau Turborepo.

## Consequences

### Positif
- Satu PR bisa update API spec + backend handler + TS client + UI sekaligus
  → tidak ada drift schema
- Versioning sederhana (single source of truth)
- Code review lebih konteks (reviewer lihat full slice)
- Mobile (phase 2) bisa langsung re-use `packages/api-client` & `packages/types`
  → minimal perubahan

### Negatif
- Repo membesar seiring waktu
- CI/CD perlu path-based filter agar build cepat
- Permission management lebih kompleks kalau team membesar

## Alternatives Considered

### A. Multi-repo (`setorin-backend`, `setorin-frontend`, `setorin-mobile`)
- **Ditolak**: schema drift jadi masalah serius di tim kecil. Versi API client
  harus selalu di-bump manual. Cross-cutting change (mis. tambah field) butuh
  3 PR di 3 repo.

### B. Monorepo tapi backend dipisah ke `setorin-backend`
- **Ditolak**: kontradiktif. Kalau monorepo untuk frontend artinya tetap satu,
  kenapa backend dipisah? Trade-off-nya tidak worth it.

## Migration Path
Kalau di kemudian hari (>10 engineers, atau backend butuh release cycle berbeda)
ingin pisah repo: backend bisa di-extract ke repo baru karena `backend/` sudah
self-contained (`go.mod` sendiri, tidak depend ke parent).
