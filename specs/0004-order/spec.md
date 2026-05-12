# Spec 0004 — Order Management (Cash Payment)

**Status:** Draft → In Implementation
**Last updated:** 2026-05-11
**Owner:** @ekohk
**Depends on:** Spec 0002 (auth), Spec 0003 (catalog)

## 1. Problem
User butuh cara konkret menjual sampah ke collector. Collector butuh dashboard
menerima/memproses order. Pembayaran **cash on hand** saat driver datang
(MVP), wallet/digital nanti (Phase 2).

## 2. Goals
| # | Goal | Acceptance |
|---|---|---|
| G1 | User bisa buat order (pilih material, weight estimate, foto, method) | POST /v1/orders return 201 + order_code (ECC-XXXXX) |
| G2 | Collector lihat order incoming di area mereka | GET /v1/collector/orders/incoming return list pending |
| G3 | 8-stage state machine ter-validate (no skip stage) | Endpoint reject invalid transition |
| G4 | OTP 4-digit verifikasi serah terima | Generated saat accept, di-verify saat arrive |
| G5 | Quality grading + bonus | Admin set bonus % per grade (hardcoded MVP) |
| G6 | User confirm cash received → order COMPLETE | Final stage requires user action |
| G7 | Email notif tiap stage transition penting | Order created/accepted/enroute/arrived/done |

## 3. Non-Goals (v1)
- ❌ Wallet/digital payment — Phase 2 (sudah archived spec)
- ❌ Real-time driver tracking GPS — Phase 2
- ❌ In-app chat user ↔ driver — Phase 2 (pakai WhatsApp deeplink dulu)
- ❌ Auto-assign collector — user pilih dari Map
- ❌ Multi-material per order — 1 material per order untuk MVP
- ❌ Refund flow — Phase 2

## 4. State Machine

```
                ┌──────────┐
                │ received │  ← user create order
                └────┬─────┘
                     │ collector ACCEPT (generate OTP)
                     │ user CANCEL → cancelled
                     ▼
                ┌──────────┐
                │ accepted │
                └────┬─────┘
                     │ driver START PICKUP / DROPOFF
                     ▼
                ┌──────────┐
                │ enroute  │
                └────┬─────┘
                     │ driver ARRIVE
                     ▼
                ┌──────────┐
                │ arrived  │  ← user shows OTP, driver verifies
                └────┬─────┘
                     │ OTP verified → next
                     ▼
                ┌──────────┐
                │ weighing │  ← driver inputs actual_weight
                └────┬─────┘
                     │
                     ▼
                ┌──────────┐
                │ quality  │  ← driver inputs grade + checks
                └────┬─────┘
                     │ driver hands cash → next
                     ▼
                ┌────────────────┐
                │ cash_handover  │  ← waiting user confirm received cash
                └────┬───────────┘
                     │ user CONFIRM RECEIVED
                     ▼
                ┌──────────┐
                │   done   │  (terminal)
                └──────────┘

Terminal states: cancelled (anytime by user before accepted),
                 done (success), disputed (any time, escalate to admin)
```

## 5. Domain Model

### `orders`
```
id              UUID PK
order_code      VARCHAR(20) UK ('ECC-04827') ← human-friendly
user_id         UUID FK → users.id          (seller)
collector_id    UUID FK → users.id NULLABLE (assigned only after accepted)
material_id     UUID FK → materials.id
estimated_weight_kg   NUMERIC(10,3)         (e.g. 3.500)
actual_weight_kg      NUMERIC(10,3) NULLABLE
unit_price_at_order   BIGINT                (snapshot price/kg saat order, lock harga)
estimated_payout      BIGINT                (snapshot estimasi)
quality_grade         VARCHAR(2) NULLABLE   ('A','B','C')
quality_bonus_pct     INTEGER DEFAULT 0     (3 untuk grade A misal)
final_payout          BIGINT NULLABLE       (actual_weight * unit_price * (1+bonus%))
method                VARCHAR(20)           ('pickup' or 'dropoff')
payment_method        VARCHAR(20) DEFAULT 'cash'   ← Phase 2 future-proof
payment_status        VARCHAR(20) DEFAULT 'pending'
status                order_status enum
otp_code              VARCHAR(4) NULLABLE   (set saat accepted, cleared saat verified)
otp_verified_at       TIMESTAMPTZ NULLABLE
address_text          TEXT                  (snapshot dari user.address atau input)
notes                 TEXT NULLABLE         (catatan user untuk collector)
created_at, updated_at, accepted_at, arrived_at, completed_at, cancelled_at, cancelled_by, cancellation_reason
```

### `order_status_history`
Append-only audit tiap transition. (id, order_id, from_status, to_status, changed_by, notes, changed_at)

### `order_photos`
Upload foto dari user (bukti barang) & dari collector (bukti timbang).
(id, order_id, url, type ['user_before' | 'collector_weighing' | 'collector_quality'], uploaded_by, uploaded_at)

## 6. API Endpoints

### User-side
| Method | Path | Purpose |
|---|---|---|
| POST | `/v1/orders` | Create order |
| GET | `/v1/orders/me` | List my orders |
| GET | `/v1/orders/:code` | Detail (by order_code) |
| POST | `/v1/orders/:code/cancel` | Cancel (only if status=received) |
| POST | `/v1/orders/:code/confirm-cash` | Confirm received cash → done |

### Collector-side (require `collector` role)
| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/collector/orders/incoming` | List nearby pending orders |
| GET | `/v1/collector/orders/me` | List orders assigned to me |
| POST | `/v1/collector/orders/:code/accept` | Accept (generate OTP) |
| POST | `/v1/collector/orders/:code/start-pickup` | Driver start trip |
| POST | `/v1/collector/orders/:code/arrive` | Driver tiba |
| POST | `/v1/collector/orders/:code/verify-otp` | Submit OTP from user |
| POST | `/v1/collector/orders/:code/weigh` | Input actual_weight |
| POST | `/v1/collector/orders/:code/quality` | Input grade + checks |
| POST | `/v1/collector/orders/:code/handover` | Mark cash given (waiting user confirm) |

### Admin
| Method | Path | Purpose |
|---|---|---|
| GET | `/v1/admin/orders?status=...` | List all (paginated) |
| POST | `/v1/admin/orders/:code/dispute` | Escalate to disputed |

## 7. Email Notifications (Cash Flow)

Tiap stage trigger email ke user (sometimes ke collector juga):

| Stage transition | Email ke User | Email ke Collector |
|---|---|---|
| received | "Order Anda dibuat" | (none, queued di list) |
| accepted | "Collector menerima order Anda" + OTP | (none) |
| enroute | "Driver dalam perjalanan" | — |
| arrived | "Driver sudah tiba" | — |
| weighing | (skip — too granular) | — |
| done | "Order selesai, terima kasih telah recycle" | — |
| cancelled | "Order dibatalkan" | "User membatalkan order" |

## 8. Open Questions
- [x] Cash flow vs wallet? **Cash MVP** (decided 2026-05-11)
- [x] OTP length? **4 digit** (sesuai design)
- [x] OTP expire? **30 min setelah accept**
- [x] Order code format? **ECC-XXXXX** (5-digit padded)
- [x] Multi-material? **Single material per order**, simpler validation
- [x] User cancel kapan saja? **Hanya saat status=received** (sebelum collector accept)

## 9. Security & Anti-Fraud
- OTP cuma 4 digit tapi expire cepat (30min) + di-clear setelah verified
- Quality grade ditentukan collector (subjective) — phase 2: foto bukti + admin review
- Final payout dihitung backend (frontend hanya display) — collector tidak bisa manipulate
- Audit trail: tiap stage di `order_status_history`
- Dispute flow → flag `status='disputed'`, admin tangani manual

## 10. Success Criteria
- 100% state transition valid (no order skip stage)
- 0 race condition saat 2 collector simultaneous accept (DB constraint + tx)
- p95 `/v1/orders/me` < 50ms
- Email landing di MailHog dalam <5 detik
