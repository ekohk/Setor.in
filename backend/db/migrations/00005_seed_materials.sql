-- +goose Up
-- +goose StatementBegin

-- Initial 8 material categories with starting prices.
-- Admin can update via /v1/admin/materials/:id/prices after launch.
-- Prices in rupiah per kg, sourced from EcoCycle design bundle (May 2026).

WITH inserted AS (
  INSERT INTO materials (slug, name, icon, sort_order, description) VALUES
    ('plastic',   'Plastik',   'plastic',   1, 'Botol PET, kemasan HDPE, plastik bersih'),
    ('cardboard', 'Kardus',    'cardboard', 2, 'Kardus bekas, karton tebal'),
    ('paper',     'Kertas',    'paper',     3, 'Kertas HVS, koran, majalah'),
    ('aluminum',  'Aluminium', 'aluminum',  4, 'Kaleng minuman, panci, peralatan dapur'),
    ('copper',    'Tembaga',   'copper',    5, 'Kabel tembaga, pipa, komponen elektronik'),
    ('steel',     'Besi',      'steel',     6, 'Besi rongsok, baja ringan'),
    ('glass',     'Kaca',      'glass',     7, 'Botol kaca, pecahan kaca bersih'),
    ('ewaste',    'E-waste',   'ewaste',    8, 'Elektronik bekas, baterai, sirkuit')
  RETURNING id, slug
)
INSERT INTO material_prices (material_id, price_per_kg, note)
SELECT
  i.id,
  CASE i.slug
    WHEN 'plastic'   THEN 5500
    WHEN 'cardboard' THEN 2400
    WHEN 'paper'     THEN 3200
    WHEN 'aluminum'  THEN 18500
    WHEN 'copper'    THEN 92000
    WHEN 'steel'     THEN 6800
    WHEN 'glass'     THEN 1200
    WHEN 'ewaste'    THEN 22000
  END,
  'Initial seed price'
FROM inserted i;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM material_prices WHERE note = 'Initial seed price';
DELETE FROM materials WHERE slug IN
  ('plastic','cardboard','paper','aluminum','copper','steel','glass','ewaste');
-- +goose StatementEnd
