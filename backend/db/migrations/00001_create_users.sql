-- +goose Up
-- +goose StatementBegin

-- Required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Enums
CREATE TYPE user_role AS ENUM ('user', 'collector', 'admin', 'super_admin');

CREATE TYPE user_status AS ENUM (
  'pending_verification',  -- baru register, belum verify email
  'active',                -- normal user, email verified
  'suspended',             -- di-suspend admin (sementara)
  'deleted'                -- soft deleted
);

-- Main users table
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  keycloak_id     UUID NOT NULL UNIQUE,
  email           CITEXT NOT NULL UNIQUE,
  phone           VARCHAR(20) UNIQUE,
  full_name       VARCHAR(120) NOT NULL,
  avatar_url      TEXT,
  primary_role    user_role NOT NULL DEFAULT 'user',
  status          user_status NOT NULL DEFAULT 'pending_verification',
  email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
  last_login_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_keycloak_id ON users (keycloak_id);
CREATE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_status ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_primary_role ON users (primary_role) WHERE deleted_at IS NULL;

-- updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- Role history (audit perubahan role)
CREATE TABLE user_role_history (
  id          BIGSERIAL PRIMARY KEY,
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  from_role   user_role,
  to_role     user_role NOT NULL,
  changed_by  UUID REFERENCES users(id) ON DELETE SET NULL,
  reason      TEXT,
  changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_role_history_user ON user_role_history (user_id, changed_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_role_history;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
