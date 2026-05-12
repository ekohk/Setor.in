// Package ports defines outbound interfaces for the catalog module.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
)

// MaterialRepository abstracts material + price persistence.
type MaterialRepository interface {
	// ListActiveWithPrice returns active materials joined with their current price.
	// Sorted by sort_order ASC. Used by public /v1/materials.
	ListActiveWithPrice(ctx context.Context) ([]model.MaterialWithPrice, error)

	// ListAllWithPrice returns ALL materials (incl. inactive). Admin endpoint.
	ListAllWithPrice(ctx context.Context) ([]model.MaterialWithPrice, error)

	// FindBySlug returns a single material + current price by slug.
	FindBySlug(ctx context.Context, slug string) (*model.MaterialWithPrice, error)

	// FindByID returns a material by UUID (admin only — they have the ID).
	FindByID(ctx context.Context, id uuid.UUID) (*model.MaterialWithPrice, error)

	// UpdateMaterial patches name / icon / is_active / sort_order / description.
	UpdateMaterial(ctx context.Context, id uuid.UUID, p MaterialPatch) (*model.Material, error)

	// InsertPrice appends a new price row. ValidFrom defaults to NOW() if zero.
	InsertPrice(ctx context.Context, materialID uuid.UUID, pricePerKg int64, createdBy *uuid.UUID, note *string) (*model.MaterialPrice, error)

	// ListPriceHistory returns all prices for a material, newest first.
	ListPriceHistory(ctx context.Context, materialID uuid.UUID, limit int) ([]model.MaterialPrice, error)
}

// MaterialPatch carries optional updates. Nil pointer = leave unchanged.
type MaterialPatch struct {
	Name        *string
	Icon        *string
	Description *string
	IsActive    *bool
	SortOrder   *int
}

// HasAny returns true if at least one field is set.
func (p MaterialPatch) HasAny() bool {
	return p.Name != nil || p.Icon != nil || p.Description != nil ||
		p.IsActive != nil || p.SortOrder != nil
}
