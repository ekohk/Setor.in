-- +goose Up
-- +goose StatementBegin

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'cv';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL tidak mendukung DROP VALUE untuk enum secara langsung.
-- Rollback manual diperlukan jika mau menghapus role ini.
-- +goose StatementEnd