# Spec 0003 — Catalog & Pricing

**Status:** Draft → In Implementation
**Last updated:** 2026-05-11
**Owner:** @ekohk
**Depends on:** Spec 0002 (auth — admin role required for price updates)

## 1. Problem
User di Home screen butuh lihat **harga material harian** untuk tahu nilai
sampah mereka sebelum jual. Admin perlu cara update harga harian sesuai
fluktuasi pasar (harga komoditas naik-turun).

## 2. Goals
| # | Goal | Acceptance |
|---|---|---|
| G1 | 8 material kategori tersedia sebagai master data | Migration seed insert 8 row |
| G2 | Harga material time-series (admin update tanpa kehilangan history) | Tabel `material_prices` append-only |
| G3 | Frontend dapat list material + harga aktif dengan 1 request | `GET /v1/materials` return array dengan `unit_price` di-join |
| G4 | Admin bisa update harga tanpa down-time | POST harga baru = harga aktif effective `valid_from` |
| G5 | History harga ter-track untuk analytics/audit | `GET /v1/admin/materials/:id/prices` return semua entry |

## 3. Non-Goals (v1)
- ❌ Auto-pull harga dari API eksternal (Bappebti, dll) — phase 2
- ❌ Harga per-region/per-kota — semua harga nasional dulu
- ❌ Harga per-tier collector (premium vs reguler) — phase 2
- ❌ Discount/promo material — out of scope
- ❌ Multi-currency — Rupiah only

## 4. Domain Model

```
materials                          material_prices
─────────                          ────────────────
id (UUID PK)                       id (UUID PK)
slug (UNIQUE)        ─────────────→ material_id (FK)
name                               price_per_kg (BIGINT, in rupiah)
icon                               valid_from (TIMESTAMPTZ)
unit (default 'kg')                created_by (FK users.id)
description                        created_at
is_active (bool)
sort_order
created_at
updated_at
```

**Aturan harga aktif:** harga aktif = harga dengan `valid_from <= NOW()`
terbaru per `material_id`. Tidak ada flag "current" — derive dari query.

## 5. API Endpoints

### Public (no auth needed)
| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/materials` | List active materials with current price |
| GET | `/v1/materials/:slug` | Detail one material |

### Admin
| Method | Path | Role | Purpose |
|---|---|---|---|
| GET | `/v1/admin/materials` | admin | List all (incl. inactive) |
| PATCH | `/v1/admin/materials/:id` | admin | Update name/icon/is_active/sort_order |
| GET | `/v1/admin/materials/:id/prices` | admin | Price history |
| POST | `/v1/admin/materials/:id/prices` | admin | Set new price (append-only) |

## 6. Initial 8 Materials (seed)

Mengikuti [icons.jsx](../../project/icons.jsx#L78-L87) dari design bundle:

| Slug | Name | Initial Price (Rp/kg) |
|---|---|---|
| plastic | Plastik | 5500 |
| cardboard | Kardus | 2400 |
| paper | Kertas | 3200 |
| aluminum | Aluminium | 18500 |
| copper | Tembaga | 92000 |
| steel | Besi | 6800 |
| glass | Kaca | 1200 |
| ewaste | E-waste | 22000 |

Initial price diset dari migration seed. Admin bisa update via API setelahnya.

## 7. Open Questions (resolved)

- [x] Harga manual vs API external? **Manual (admin) untuk MVP**
- [x] Harga global atau per-region? **Global**
- [x] Harga aktif disimpan denormalized atau di-derive? **Di-derive (subquery LATERAL)**
- [x] Allow delete harga history? **Tidak — append-only audit trail**

## 8. Success Criteria

- `GET /v1/materials` p95 < 50ms (8 rows, simple LATERAL join)
- Admin update harga, frontend baca harga baru dalam request berikutnya (no caching MVP)
- 0 race condition saat 2 admin update harga sekaligus (timestamp-based, last-write-wins)
- Material list ter-sort sesuai `sort_order` (admin atur urutan tampil di frontend)
