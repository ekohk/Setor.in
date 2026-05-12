-- +goose Up
-- +goose StatementBegin

CREATE TABLE materials (
  id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  slug         VARCHAR(40) NOT NULL UNIQUE,
  name         VARCHAR(80) NOT NULL,
  icon         VARCHAR(40),                          -- icon slug for frontend (e.g. 'plastic', 'copper')
  unit         VARCHAR(10) NOT NULL DEFAULT 'kg',
  description  TEXT,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order   INTEGER NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_materials_active_sort
  ON materials (is_active, sort_order)
  WHERE is_active = TRUE;

CREATE TRIGGER trg_materials_updated_at
  BEFORE UPDATE ON materials
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- Price history. Append-only. The "active" price for a material is the row
-- with the latest valid_from <= NOW().
CREATE TABLE material_prices (
  id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  material_id   UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
  price_per_kg  BIGINT NOT NULL CHECK (price_per_kg > 0),    -- in rupiah, integer (no fractional rupiah)
  valid_from    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
  note          TEXT,                                         -- optional admin note ("market spike", etc.)
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hot path: lookup current price for a material
CREATE INDEX idx_material_prices_lookup
  ON material_prices (material_id, valid_from DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS material_prices;
DROP TRIGGER IF EXISTS trg_materials_updated_at ON materials;
DROP TABLE IF EXISTS materials;
-- +goose StatementEnd
