# Spec 0004 — Wallet (Internal Ledger)

**Status:** Draft → In Implementation (Phase 1)
**Last updated:** 2026-05-11
**Owner:** @ekohk
**Depends on:** Spec 0002 (users)

## 1. Problem
Order yang sukses harus credit ke user (penjual). Butuh tempat menyimpan saldo
yang **akurat finansial**, **auditable**, dan **siap diintegrasikan ke
payment gateway** untuk withdraw nanti.

## 2. Goals
| # | Goal | Acceptance |
|---|---|---|
| G1 | Setiap user punya 1 wallet otomatis sejak first sync | INSERT wallet row dalam transaksi sync |
| G2 | Saldo akurat (no race condition) | `SELECT FOR UPDATE` saat mutate; pgx atomic tx |
| G3 | Audit trail lengkap | `wallet_transactions` append-only, idempotent via reference |
| G4 | Withdraw request bisa di-create, di-cancel, status tracking | Withdrawal tabel + state `pending/processing/paid/failed/cancelled` |
| G5 | Saldo tidak bisa negatif | DB CHECK + app-level guard |
| G6 | Integer rupiah, no fractional | BIGINT amount, no NUMERIC/FLOAT |

## 3. Non-Goals (Phase 1)
- ❌ Integrasi gateway real (Xendit, Flip) — Phase 2
- ❌ Top-up dari bank — out of scope
- ❌ Transfer antar user — out of scope
- ❌ Multi-currency — Rupiah only
- ❌ Pending balance / hold logic — Phase 2 (untuk now, payout langsung available)

## 4. Domain Model

```
wallets                              wallet_transactions
─────────                            ─────────────────────
id (UUID PK)                         id (UUID PK)
user_id (UUID UK, FK)  ─────────→    wallet_id (FK)
balance (BIGINT)                     type (enum)
                                     amount (BIGINT, always positive)
                                     direction (enum: credit/debit)
                                     reference_type (varchar)
                                     reference_id (varchar)
                                     description
                                     balance_before (BIGINT)
                                     balance_after (BIGINT)
                                     created_by (FK users.id)
                                     created_at
                                     UNIQUE(reference_type, reference_id)

withdrawals
─────────────
id (UUID PK)
wallet_id (FK)
user_id (FK)
amount (BIGINT)
fee (BIGINT)
bank_code (varchar)
bank_account_number (varchar) -- nanti encrypt at rest
bank_account_holder (varchar)
status (enum)
external_ref (varchar)  -- gateway tx id, set di Phase 2
requested_at
processed_at
paid_at
failed_reason
```

## 5. Transaction Types (`wallet_transactions.type`)
- `order_payout` — credit dari order yang selesai
- `withdrawal` — debit saat user request withdraw
- `withdrawal_refund` — credit balik kalau withdraw gagal
- `bonus` — credit referral / promo
- `adjustment` — manual oleh admin (debit atau credit), wajib note
- `correction` — koreksi balance setelah audit

## 6. State Machine: Withdrawal

```
                  ┌─────────┐
                  │ pending │  ←── user create
                  └────┬────┘
       ┌───────────────┼─────────────────┐
       │ cancel        │ approve          │ expire (auto 24h)
       ▼               ▼                  ▼
  ┌─────────┐    ┌────────────┐     ┌──────────┐
  │cancelled│    │ processing │     │ expired  │
  └─────────┘    └─────┬──────┘     └──────────┘
                       │
              ┌────────┴────────┐
              │                 │
              ▼                 ▼
         ┌────────┐        ┌────────┐
         │  paid  │        │ failed │
         └────────┘        └────┬───┘
                                │ admin refund
                                ▼
                            ┌──────────┐
                            │ refunded │
                            └──────────┘
```

Phase 1: hanya `pending` & `cancelled` & `expired` aktif. `processing/paid/failed` butuh gateway.

## 7. API Endpoints (Phase 1)

| Method | Path | Role | Purpose |
|---|---|---|---|
| GET | `/v1/wallet/me` | user | Saldo + ringkasan |
| GET | `/v1/wallet/me/transactions?page=` | user | History (paginated) |
| POST | `/v1/wallet/me/withdrawals` | user | Request withdraw (pending) |
| GET | `/v1/wallet/me/withdrawals` | user | List my withdrawals |
| POST | `/v1/wallet/me/withdrawals/:id/cancel` | user | Cancel pending |
| GET | `/v1/admin/withdrawals` | admin | List all (audit) |
| GET | `/v1/admin/wallets/:user_id` | admin | Inspect a user's wallet |
| POST | `/v1/admin/wallets/:user_id/adjust` | super_admin | Manual adjust dengan note |

## 8. Atomic Operations

Semua mutate saldo WAJIB pakai `pgxpool.BeginTx()`:

```go
// Pseudocode
tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
defer tx.Rollback(ctx)

// 1. Lock wallet row
SELECT id, balance FROM wallets WHERE id = $1 FOR UPDATE

// 2. Compute new balance
newBalance := balance + amount  // or - amount for debit

// 3. Guard: no negative
if newBalance < 0 { return ErrInsufficientFunds }

// 4. Update balance
UPDATE wallets SET balance = $1 WHERE id = $2

// 5. Insert ledger row
INSERT INTO wallet_transactions (...) VALUES (...)
    ON CONFLICT (reference_type, reference_id) DO NOTHING
RETURNING id

// 6. Commit
tx.Commit(ctx)
```

## 9. Security Requirements
- [x] No `UPDATE balance = balance + X` outside `BeginTx` — enforced by code review
- [x] Idempotency via `(reference_type, reference_id)` unique constraint
- [x] DB-level `CHECK (balance >= 0)` constraint
- [x] All financial endpoints rate-limited (TODO: middleware in Phase 2)
- [x] Audit: every mutate logs to slog + `auth_events`
- [x] Bank account number encryption at rest (TODO: Phase 2, when withdrawal goes live)

## 10. Open Questions (resolved)

- [x] Wallet auto-create di sync? **Yes**
- [x] Allow negative balance? **No (DB CHECK)**
- [x] Withdrawal fee model? **Flat Rp 2.500 untuk MVP, configurable nanti**
- [x] Min/max withdrawal? **Min Rp 10.000, max harian Rp 5jt untuk MVP**

## 11. Success Criteria
- 100% balance accurate setelah 1000 concurrent operations (load test)
- 0 transaksi double-credit dengan duplicate reference
- p95 `/v1/wallet/me` < 30ms
- Audit log lengkap: setiap mutate ada row di `wallet_transactions`
