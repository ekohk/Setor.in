-- +goose Up
-- +goose StatementBegin

-- PostGIS optional di sini, untuk geo-search collector nanti.
-- Kalau extension belum ada, install via: CREATE EXTENSION postgis;
CREATE EXTENSION IF NOT EXISTS "postgis";

CREATE TYPE application_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE collector_applications (
  id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  business_name     VARCHAR(120) NOT NULL,
  license_no        VARCHAR(40),
  ktp_url           TEXT NOT NULL,
  siup_url          TEXT,
  address           TEXT NOT NULL,
  location          GEOGRAPHY(POINT, 4326),
  status            application_status NOT NULL DEFAULT 'pending',
  rejection_reason  TEXT,
  reviewed_by       UUID REFERENCES users(id) ON DELETE SET NULL,
  submitted_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reviewed_at       TIMESTAMPTZ
);

CREATE INDEX idx_collector_apps_user ON collector_applications (user_id);
CREATE INDEX idx_collector_apps_status ON collector_applications (status, submitted_at DESC);
CREATE INDEX idx_collector_apps_location
  ON collector_applications USING GIST (location)
  WHERE location IS NOT NULL;

-- Hanya 1 application aktif (pending) per user
CREATE UNIQUE INDEX idx_one_pending_application
  ON collector_applications (user_id)
  WHERE status = 'pending';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collector_applications;
DROP TYPE IF EXISTS application_status;
-- Tidak DROP EXTENSION postgis — bisa dipakai modul lain
-- +goose StatementEnd
