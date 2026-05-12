// Package dto holds request/response shapes for the catalog module.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
)

// MaterialResponse is the public-facing shape of a Material (with current price).
type MaterialResponse struct {
	ID             uuid.UUID  `json:"id"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Icon           *string    `json:"icon,omitempty"`
	Unit           string     `json:"unit"`
	Description    *string    `json:"description,omitempty"`
	IsActive       bool       `json:"is_active"`
	SortOrder      int        `json:"sort_order"`
	CurrentPrice   *int64     `json:"current_price,omitempty"`   // rupiah/kg, nil if not yet set
	PriceValidFrom *time.Time `json:"price_valid_from,omitempty"`
}

// FromMaterialWithPrice converts a domain row to the API shape.
func FromMaterialWithPrice(m *model.MaterialWithPrice) MaterialResponse {
	return MaterialResponse{
		ID:             m.ID,
		Slug:           m.Slug,
		Name:           m.Name,
		Icon:           m.Icon,
		Unit:           m.Unit,
		Description:    m.Description,
		IsActive:       m.IsActive,
		SortOrder:      m.SortOrder,
		CurrentPrice:   m.CurrentPrice,
		PriceValidFrom: m.PriceValidFrom,
	}
}

// MaterialPriceResponse is one entry in price history.
type MaterialPriceResponse struct {
	ID         uuid.UUID  `json:"id"`
	PricePerKg int64      `json:"price_per_kg"`
	ValidFrom  time.Time  `json:"valid_from"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
	Note       *string    `json:"note,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func FromMaterialPrice(p *model.MaterialPrice) MaterialPriceResponse {
	return MaterialPriceResponse{
		ID:         p.ID,
		PricePerKg: p.PricePerKg,
		ValidFrom:  p.ValidFrom,
		CreatedBy:  p.CreatedBy,
		Note:       p.Note,
		CreatedAt:  p.CreatedAt,
	}
}

// UpdateMaterialRequest is the body for PATCH /v1/admin/materials/:id.
type UpdateMaterialRequest struct {
	Name        *string `json:"name,omitempty"        binding:"omitempty,min=1,max=80"`
	Icon        *string `json:"icon,omitempty"        binding:"omitempty,max=40"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"  binding:"omitempty,min=0,max=999"`
}

// SetPriceRequest is the body for POST /v1/admin/materials/:id/prices.
type SetPriceRequest struct {
	PricePerKg int64   `json:"price_per_kg" binding:"required,min=1"`
	Note       *string `json:"note,omitempty"`
}
