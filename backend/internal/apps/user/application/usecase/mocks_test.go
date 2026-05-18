package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
)

// ─── MockUserRepository ──────────────────────────────────────────────────────

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByKeycloakID(ctx context.Context, kcID uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, kcID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, p ports.ProfilePatch) (*model.User, error) {
	args := m.Called(ctx, id, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) UpdateAuthState(ctx context.Context, id uuid.UUID, emailVerified bool, status model.Status) (*model.User, error) {
	args := m.Called(ctx, id, emailVerified, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// ─── MockNotifier ────────────────────────────────────────────────────────────

type MockNotifier struct{ mock.Mock }

func (m *MockNotifier) Welcome(to, fullName string) {
	m.Called(to, fullName)
}
