// Package usecase orchestrates business logic for the catalog module.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/catalog/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
)

type MaterialUseCase struct {
	repo ports.MaterialRepository
}

func NewMaterialUseCase(repo ports.MaterialRepository) *MaterialUseCase {
	return &MaterialUseCase{repo: repo}
}

// ─── Public (no auth) ───

func (uc *MaterialUseCase) ListPublic(ctx context.Context) ([]model.MaterialWithPrice, error) {
	return uc.repo.ListActiveWithPrice(ctx)
}

func (uc *MaterialUseCase) GetBySlug(ctx context.Context, slug string) (*model.MaterialWithPrice, error) {
	return uc.repo.FindBySlug(ctx, slug)
}

// ─── Admin ───

func (uc *MaterialUseCase) ListAdmin(ctx context.Context) ([]model.MaterialWithPrice, error) {
	return uc.repo.ListAllWithPrice(ctx)
}

func (uc *MaterialUseCase) GetByID(ctx context.Context, id uuid.UUID) (*model.MaterialWithPrice, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *MaterialUseCase) UpdateMaterial(ctx context.Context, id uuid.UUID, p ports.MaterialPatch) (*model.Material, error) {
	return uc.repo.UpdateMaterial(ctx, id, p)
}

// SetPrice appends a new price row. Frontend sees the new price immediately
// (no cache yet). Old prices remain in history for audit.
func (uc *MaterialUseCase) SetPrice(ctx context.Context, materialID uuid.UUID, pricePerKg int64, actorID *uuid.UUID, note *string) (*model.MaterialPrice, error) {
	return uc.repo.InsertPrice(ctx, materialID, pricePerKg, actorID, note)
}

func (uc *MaterialUseCase) PriceHistory(ctx context.Context, materialID uuid.UUID, limit int) ([]model.MaterialPrice, error) {
	return uc.repo.ListPriceHistory(ctx, materialID, limit)
}
