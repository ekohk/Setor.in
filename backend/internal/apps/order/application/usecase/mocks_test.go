package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	catalogmodel "github.com/setorin/setorin/backend/internal/apps/catalog/domain/model"
	catalogports "github.com/setorin/setorin/backend/internal/apps/catalog/application/ports"
	ordermodel "github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	orderports "github.com/setorin/setorin/backend/internal/apps/order/application/ports"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
)

// ─── MockOrderRepository ─────────────────────────────────────────────────────

type MockOrderRepository struct{ mock.Mock }

func (m *MockOrderRepository) NextOrderCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockOrderRepository) Create(ctx context.Context, o *ordermodel.Order) (*ordermodel.Order, error) {
	args := m.Called(ctx, o)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordermodel.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*ordermodel.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordermodel.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByCode(ctx context.Context, code string) (*ordermodel.Order, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordermodel.Order), args.Error(1)
}

func (m *MockOrderRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]ordermodel.Order, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	return args.Get(0).([]ordermodel.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) ListByCollector(ctx context.Context, collectorID uuid.UUID, page, pageSize int) ([]ordermodel.Order, int64, error) {
	args := m.Called(ctx, collectorID, page, pageSize)
	return args.Get(0).([]ordermodel.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) ListIncoming(ctx context.Context, page, pageSize int) ([]ordermodel.Order, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]ordermodel.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) ListAdmin(ctx context.Context, statusFilter *ordermodel.Status, page, pageSize int) ([]ordermodel.Order, int64, error) {
	args := m.Called(ctx, statusFilter, page, pageSize)
	return args.Get(0).([]ordermodel.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) Transition(ctx context.Context, orderID uuid.UUID, p orderports.TransitionParams) (*ordermodel.Order, error) {
	args := m.Called(ctx, orderID, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ordermodel.Order), args.Error(1)
}

func (m *MockOrderRepository) StatusHistory(ctx context.Context, orderID uuid.UUID) ([]ordermodel.StatusHistory, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]ordermodel.StatusHistory), args.Error(1)
}

// ─── MockMaterialRepository ──────────────────────────────────────────────────

type MockMaterialRepository struct{ mock.Mock }

func (m *MockMaterialRepository) ListActiveWithPrice(ctx context.Context) ([]catalogmodel.MaterialWithPrice, error) {
	args := m.Called(ctx)
	return args.Get(0).([]catalogmodel.MaterialWithPrice), args.Error(1)
}

func (m *MockMaterialRepository) ListAllWithPrice(ctx context.Context) ([]catalogmodel.MaterialWithPrice, error) {
	args := m.Called(ctx)
	return args.Get(0).([]catalogmodel.MaterialWithPrice), args.Error(1)
}

func (m *MockMaterialRepository) FindBySlug(ctx context.Context, slug string) (*catalogmodel.MaterialWithPrice, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogmodel.MaterialWithPrice), args.Error(1)
}

func (m *MockMaterialRepository) FindByID(ctx context.Context, id uuid.UUID) (*catalogmodel.MaterialWithPrice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogmodel.MaterialWithPrice), args.Error(1)
}

func (m *MockMaterialRepository) UpdateMaterial(ctx context.Context, id uuid.UUID, p catalogports.MaterialPatch) (*catalogmodel.Material, error) {
	args := m.Called(ctx, id, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogmodel.Material), args.Error(1)
}

func (m *MockMaterialRepository) InsertPrice(ctx context.Context, materialID uuid.UUID, pricePerKg int64, createdBy *uuid.UUID, note *string) (*catalogmodel.MaterialPrice, error) {
	args := m.Called(ctx, materialID, pricePerKg, createdBy, note)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*catalogmodel.MaterialPrice), args.Error(1)
}

func (m *MockMaterialRepository) ListPriceHistory(ctx context.Context, materialID uuid.UUID, limit int) ([]catalogmodel.MaterialPrice, error) {
	args := m.Called(ctx, materialID, limit)
	return args.Get(0).([]catalogmodel.MaterialPrice), args.Error(1)
}

// ─── MockUserRepository (for order usecase) ──────────────────────────────────

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, u *usermodel.User) (*usermodel.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}

func (m *MockUserRepository) FindByKeycloakID(ctx context.Context, kcID uuid.UUID) (*usermodel.User, error) {
	args := m.Called(ctx, kcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*usermodel.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*usermodel.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, p userports.ProfilePatch) (*usermodel.User, error) {
	args := m.Called(ctx, id, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) UpdateAuthState(ctx context.Context, id uuid.UUID, emailVerified bool, status usermodel.Status) (*usermodel.User, error) {
	args := m.Called(ctx, id, emailVerified, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usermodel.User), args.Error(1)
}
