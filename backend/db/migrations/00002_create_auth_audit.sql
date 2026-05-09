-- +goose Up
-- +goose StatementBegin

-- Audit trail untuk semua event terkait auth.
-- Append-only. Untuk security monitoring & compliance.
CREATE TABLE auth_events (
  id           BIGSERIAL PRIMARY KEY,
  user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
  keycloak_id  UUID,                  -- dipakai kalau user_id belum ada
  event_type   VARCHAR(40) NOT NULL,  -- lihat data-model.md untuk daftar
  ip_address   INET,
  user_agent   TEXT,
  metadata     JSONB,
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_events_user ON auth_events (user_id, occurred_at DESC);
CREATE INDEX idx_auth_events_type ON auth_events (event_type, occurred_at DESC);
CREATE INDEX idx_auth_events_keycloak_id ON auth_events (keycloak_id) WHERE keycloak_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_events;
-- +goose StatementEnd
