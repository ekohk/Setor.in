package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/catalog/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

type PostgresMaterialRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresMaterialRepo(pool *pgxpool.Pool) *PostgresMaterialRepo {
	return &PostgresMaterialRepo{pool: pool}
}

// Columns + LATERAL join to get current price.
// LATERAL subquery picks the latest valid_from <= NOW() row per material —
// this is the canonical "active price" query and Postgres optimizes it well
// with idx_material_prices_lookup.
const materialWithPriceCols = `
m.id, m.slug, m.name, m.icon, m.unit, m.description, m.is_active, m.sort_order,
m.created_at, m.updated_at,
mp.price_per_kg AS current_price,
mp.valid_from   AS price_valid_from`

const materialFromJoin = `
FROM materials m
LEFT JOIN LATERAL (
  SELECT price_per_kg, valid_from
  FROM material_prices
  WHERE material_id = m.id AND valid_from <= NOW()
  ORDER BY valid_from DESC
  LIMIT 1
) mp ON TRUE`

func (r *PostgresMaterialRepo) ListActiveWithPrice(ctx context.Context) ([]model.MaterialWithPrice, error) {
	q := `SELECT ` + materialWithPriceCols + materialFromJoin +
		` WHERE m.is_active = TRUE ORDER BY m.sort_order ASC, m.name ASC`
	return r.queryList(ctx, q)
}

func (r *PostgresMaterialRepo) ListAllWithPrice(ctx context.Context) ([]model.MaterialWithPrice, error) {
	q := `SELECT ` + materialWithPriceCols + materialFromJoin +
		` ORDER BY m.sort_order ASC, m.name ASC`
	return r.queryList(ctx, q)
}

func (r *PostgresMaterialRepo) FindBySlug(ctx context.Context, slug string) (*model.MaterialWithPrice, error) {
	q := `SELECT ` + materialWithPriceCols + materialFromJoin + ` WHERE m.slug = $1`
	return r.queryOne(ctx, q, slug)
}

func (r *PostgresMaterialRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.MaterialWithPrice, error) {
	q := `SELECT ` + materialWithPriceCols + materialFromJoin + ` WHERE m.id = $1`
	return r.queryOne(ctx, q, id)
}

func (r *PostgresMaterialRepo) UpdateMaterial(ctx context.Context, id uuid.UUID, p ports.MaterialPatch) (*model.Material, error) {
	if !p.HasAny() {
		// Nothing to change. Return current row.
		current, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return &current.Material, nil
	}

	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
	i := 1

	if p.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *p.Name)
		i++
	}
	if p.Icon != nil {
		sets = append(sets, fmt.Sprintf("icon = $%d", i))
		args = append(args, *p.Icon)
		i++
	}
	if p.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *p.Description)
		i++
	}
	if p.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", i))
		args = append(args, *p.IsActive)
		i++
	}
	if p.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", i))
		args = append(args, *p.SortOrder)
		i++
	}
	args = append(args, id)

	q := fmt.Sprintf(`UPDATE materials SET %s WHERE id = $%d
		RETURNING id, slug, name, icon, unit, description, is_active, sort_order, created_at, updated_at`,
		strings.Join(sets, ", "), i)

	row := r.pool.QueryRow(ctx, q, args...)
	var m model.Material
	err := row.Scan(&m.ID, &m.Slug, &m.Name, &m.Icon, &m.Unit, &m.Description,
		&m.IsActive, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "material not found")
		}
		return nil, fmt.Errorf("update material: %w", err)
	}
	return &m, nil
}

func (r *PostgresMaterialRepo) InsertPrice(ctx context.Context, materialID uuid.UUID, pricePerKg int64, createdBy *uuid.UUID, note *string) (*model.MaterialPrice, error) {
	const q = `
INSERT INTO material_prices (material_id, price_per_kg, created_by, note)
VALUES ($1, $2, $3, $4)
RETURNING id, material_id, price_per_kg, valid_from, created_by, note, created_at`

	row := r.pool.QueryRow(ctx, q, materialID, pricePerKg, createdBy, note)
	var p model.MaterialPrice
	err := row.Scan(&p.ID, &p.MaterialID, &p.PricePerKg, &p.ValidFrom, &p.CreatedBy, &p.Note, &p.CreatedAt)
	if err != nil {
		// FK violation = material doesn't exist
		if strings.Contains(err.Error(), "SQLSTATE 23503") {
			return nil, apperr.New(apperr.CodeNotFound, "material not found")
		}
		return nil, fmt.Errorf("insert price: %w", err)
	}
	return &p, nil
}

func (r *PostgresMaterialRepo) ListPriceHistory(ctx context.Context, materialID uuid.UUID, limit int) ([]model.MaterialPrice, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
SELECT id, material_id, price_per_kg, valid_from, created_by, note, created_at
FROM material_prices
WHERE material_id = $1
ORDER BY valid_from DESC
LIMIT $2`

	rows, err := r.pool.Query(ctx, q, materialID, limit)
	if err != nil {
		return nil, fmt.Errorf("list price history: %w", err)
	}
	defer rows.Close()

	out := make([]model.MaterialPrice, 0, limit)
	for rows.Next() {
		var p model.MaterialPrice
		if err := rows.Scan(&p.ID, &p.MaterialID, &p.PricePerKg, &p.ValidFrom,
			&p.CreatedBy, &p.Note, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// queryList is the shared helper for *WithPrice list queries.
func (r *PostgresMaterialRepo) queryList(ctx context.Context, q string, args ...any) ([]model.MaterialWithPrice, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query materials: %w", err)
	}
	defer rows.Close()

	out := make([]model.MaterialWithPrice, 0, 16)
	for rows.Next() {
		m, err := scanWithPrice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *PostgresMaterialRepo) queryOne(ctx context.Context, q string, args ...any) (*model.MaterialWithPrice, error) {
	row := r.pool.QueryRow(ctx, q, args...)
	m, err := scanWithPrice(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "material not found")
		}
		return nil, fmt.Errorf("query material: %w", err)
	}
	return m, nil
}

// scanWithPrice scans a row matching materialWithPriceCols order.
func scanWithPrice(row pgx.Row) (*model.MaterialWithPrice, error) {
	var m model.MaterialWithPrice
	err := row.Scan(
		&m.ID, &m.Slug, &m.Name, &m.Icon, &m.Unit, &m.Description,
		&m.IsActive, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
		&m.CurrentPrice, &m.PriceValidFrom,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
