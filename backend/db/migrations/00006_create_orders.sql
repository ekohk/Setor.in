-- +goose Up
-- +goose StatementBegin

-- Order status state machine. See specs/0004-order/spec.md §4 for diagram.
-- Terminal states: cancelled, done, disputed.
CREATE TYPE order_status AS ENUM (
  'received',
  'accepted',
  'enroute',
  'arrived',
  'weighing',
  'quality',
  'cash_handover',
  'done',
  'cancelled',
  'disputed'
);

CREATE TYPE order_method AS ENUM ('pickup', 'dropoff');

-- Sequence for human-friendly order codes (ECC-04827).
-- Padded to 5 digits via LPAD in app code.
CREATE SEQUENCE order_code_seq START 1000;

CREATE TABLE orders (
  id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_code            VARCHAR(20) NOT NULL UNIQUE,

  user_id               UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  collector_id          UUID REFERENCES users(id) ON DELETE SET NULL,
  material_id           UUID NOT NULL REFERENCES materials(id) ON DELETE RESTRICT,

  -- Weights (NUMERIC for fractional kg, e.g. 3.500). String → DB conversion handled by pgx.
  estimated_weight_kg   NUMERIC(10,3) NOT NULL CHECK (estimated_weight_kg > 0),
  actual_weight_kg      NUMERIC(10,3) CHECK (actual_weight_kg IS NULL OR actual_weight_kg > 0),

  -- Snapshot of price at time of order. Lock against price changes mid-flow.
  unit_price_at_order   BIGINT NOT NULL CHECK (unit_price_at_order > 0),
  estimated_payout      BIGINT NOT NULL CHECK (estimated_payout > 0),
  final_payout          BIGINT CHECK (final_payout IS NULL OR final_payout >= 0),

  -- Quality grading
  quality_grade         VARCHAR(2),  -- 'A', 'B', 'C'
  quality_bonus_pct     INTEGER NOT NULL DEFAULT 0 CHECK (quality_bonus_pct >= 0 AND quality_bonus_pct <= 100),

  method                order_method NOT NULL,

  -- Future-proof for wallet integration in Phase 2. For MVP always 'cash'.
  payment_method        VARCHAR(20) NOT NULL DEFAULT 'cash',
  payment_status        VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | paid | disputed
  paid_at               TIMESTAMPTZ,

  status                order_status NOT NULL DEFAULT 'received',

  -- OTP for handover verification. 4-digit string. NULL = not yet accepted or already verified.
  otp_code              VARCHAR(4),
  otp_verified_at       TIMESTAMPTZ,
  otp_expires_at        TIMESTAMPTZ,

  -- Snapshot of address at time of order (so it doesn't change if user updates profile).
  address_text          TEXT NOT NULL,
  latitude              DOUBLE PRECISION,
  longitude             DOUBLE PRECISION,
  notes                 TEXT,

  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  accepted_at           TIMESTAMPTZ,
  arrived_at            TIMESTAMPTZ,
  completed_at          TIMESTAMPTZ,
  cancelled_at          TIMESTAMPTZ,
  cancelled_by          UUID REFERENCES users(id) ON DELETE SET NULL,
  cancellation_reason   TEXT
);

CREATE INDEX idx_orders_user_status ON orders (user_id, status, created_at DESC);
CREATE INDEX idx_orders_collector_status ON orders (collector_id, status, created_at DESC) WHERE collector_id IS NOT NULL;
CREATE INDEX idx_orders_status ON orders (status, created_at DESC);
CREATE INDEX idx_orders_pending_no_collector ON orders (created_at DESC) WHERE status = 'received' AND collector_id IS NULL;

CREATE TRIGGER trg_orders_updated_at
  BEFORE UPDATE ON orders
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- Append-only audit of every status transition.
CREATE TABLE order_status_history (
  id          BIGSERIAL PRIMARY KEY,
  order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  from_status order_status,
  to_status   order_status NOT NULL,
  changed_by  UUID REFERENCES users(id) ON DELETE SET NULL,
  notes       TEXT,
  changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_status_history ON order_status_history (order_id, changed_at);

-- Photos uploaded throughout the flow.
CREATE TYPE order_photo_type AS ENUM (
  'user_before',          -- foto dari user saat buat order
  'collector_weighing',   -- foto timbangan
  'collector_quality',    -- foto material untuk quality check
  'cash_proof'            -- (optional) foto serah terima cash
);

CREATE TABLE order_photos (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  url         TEXT NOT NULL,
  type        order_photo_type NOT NULL,
  uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
  uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_photos_order ON order_photos (order_id, uploaded_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_photos;
DROP TYPE IF EXISTS order_photo_type;
DROP TABLE IF EXISTS order_status_history;
DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TABLE IF EXISTS orders;
DROP SEQUENCE IF EXISTS order_code_seq;
DROP TYPE IF EXISTS order_method;
DROP TYPE IF EXISTS order_status;
-- +goose StatementEnd
