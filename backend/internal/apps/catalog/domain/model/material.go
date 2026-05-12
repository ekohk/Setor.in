// Package model contains the catalog aggregates: Material + MaterialPrice.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Material is a recyclable category sellers can list (plastic, copper, ...).
// Master data — admin-managed. Prices live in MaterialPrice (time-series).
type Material struct {
	ID          uuid.UUID `db:"id"`
	Slug        string    `db:"slug"`        // stable lowercase identifier (e.g. "plastic")
	Name        string    `db:"name"`        // display name in Indonesian
	Icon        *string   `db:"icon"`        // icon slug for frontend
	Unit        string    `db:"unit"`        // always "kg" for v1
	Description *string   `db:"description"`
	IsActive    bool      `db:"is_active"`
	SortOrder   int       `db:"sort_order"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// MaterialPrice is one row in the price history. Append-only.
type MaterialPrice struct {
	ID          uuid.UUID  `db:"id"`
	MaterialID  uuid.UUID  `db:"material_id"`
	PricePerKg  int64      `db:"price_per_kg"` // rupiah (integer — no fractional rupiah)
	ValidFrom   time.Time  `db:"valid_from"`
	CreatedBy   *uuid.UUID `db:"created_by"` // users.id of admin who set this
	Note        *string    `db:"note"`
	CreatedAt   time.Time  `db:"created_at"`
}

// MaterialWithPrice is a denormalized view: material + its current active price.
// Returned by repository's list/lookup methods to save callers a second query.
type MaterialWithPrice struct {
	Material
	CurrentPrice    *int64     `db:"current_price"`     // nil if no price set
	PriceValidFrom  *time.Time `db:"price_valid_from"`
}
