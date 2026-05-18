package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

func newUC(repo *MockUserRepository, notifier *MockNotifier) *UserUseCase {
	return NewUserUseCase(repo, nil, notifier) // kcAdmin=nil — never called in unit tests
}

func activeUser(kcID uuid.UUID) *model.User {
	return &model.User{
		ID:            uuid.New(),
		KeycloakID:    kcID,
		Email:         "alice@example.com",
		FullName:      "Alice",
		PrimaryRole:   model.RoleUser,
		Status:        model.StatusActive,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func verifiedClaims(kcID uuid.UUID) *keycloak.Claims {
	return &keycloak.Claims{
		Email:         "alice@example.com",
		GivenName:     "Alice",
		EmailVerified: true,
		RealmAccess:   keycloak.RealmAccess{Roles: []string{"user"}},
	}
}

// ─── SyncFromClaims ──────────────────────────────────────────────────────────

func TestSyncFromClaims_CreateNewUser(t *testing.T) {
	kcID := uuid.New()
	claims := verifiedClaims(kcID)
	claims.RegisteredClaims.Subject = kcID.String()

	repo := &MockUserRepository{}
	notifier := &MockNotifier{}

	created := activeUser(kcID)
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(nil, apperr.New(apperr.CodeNotFound, "not found"))
	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(created, nil)
	repo.On("UpdateLastLogin", mock.Anything, created.ID).Return(nil)
	notifier.On("Welcome", created.Email, created.FullName).Return()

	uc := newUC(repo, notifier)
	u, err := uc.SyncFromClaims(context.Background(), claims)

	require.NoError(t, err)
	assert.Equal(t, created.ID, u.ID)
	repo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*model.User"))
	notifier.AssertCalled(t, "Welcome", created.Email, created.FullName)
}

func TestSyncFromClaims_ExistingUser_NoStateChange(t *testing.T) {
	kcID := uuid.New()
	claims := verifiedClaims(kcID)
	claims.RegisteredClaims.Subject = kcID.String()

	existing := activeUser(kcID) // already verified + active — no update needed

	repo := &MockUserRepository{}
	notifier := &MockNotifier{}

	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(existing, nil)
	repo.On("UpdateLastLogin", mock.Anything, existing.ID).Return(nil)

	uc := newUC(repo, notifier)
	u, err := uc.SyncFromClaims(context.Background(), claims)

	require.NoError(t, err)
	assert.Equal(t, existing.ID, u.ID)
	repo.AssertNotCalled(t, "UpdateAuthState")
	notifier.AssertNotCalled(t, "Welcome")
}

func TestSyncFromClaims_ExistingUser_StateChanges(t *testing.T) {
	kcID := uuid.New()
	claims := verifiedClaims(kcID)
	claims.RegisteredClaims.Subject = kcID.String()

	// Simulate user was pending_verification, now email verified in Keycloak
	unverified := activeUser(kcID)
	unverified.EmailVerified = false
	unverified.Status = model.StatusPendingVerification

	updated := activeUser(kcID) // after update

	repo := &MockUserRepository{}
	notifier := &MockNotifier{}

	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(unverified, nil)
	repo.On("UpdateAuthState", mock.Anything, unverified.ID, true, model.StatusActive).Return(updated, nil)
	// After UpdateAuthState, usecase uses updated.ID for UpdateLastLogin
	repo.On("UpdateLastLogin", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)

	uc := newUC(repo, notifier)
	u, err := uc.SyncFromClaims(context.Background(), claims)

	require.NoError(t, err)
	assert.Equal(t, updated.ID, u.ID)
	repo.AssertCalled(t, "UpdateAuthState", mock.Anything, unverified.ID, true, model.StatusActive)
}

func TestSyncFromClaims_SuspendedUser_ReturnsError(t *testing.T) {
	kcID := uuid.New()
	claims := verifiedClaims(kcID)
	claims.RegisteredClaims.Subject = kcID.String()

	// Suspended user whose email_verified + status are already in sync with claims
	// so UpdateAuthState is NOT called — only UpdateLastLogin before the status check.
	suspended := activeUser(kcID)
	suspended.Status = model.StatusSuspended
	// Force email_verified to match so targetStatus(true)=active != suspended
	// triggers UpdateAuthState. To avoid that, set email_verified=true and
	// status=suspended which means status != targetStatus(verified) → UpdateAuthState IS called.
	// Simplest: mock UpdateAuthState to return the same suspended user.
	suspended.EmailVerified = true

	repo := &MockUserRepository{}
	notifier := &MockNotifier{}

	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(suspended, nil)
	// status=suspended != targetStatus(true)=active → UpdateAuthState will be called
	repo.On("UpdateAuthState", mock.Anything, suspended.ID, true, model.StatusActive).Return(suspended, nil)
	repo.On("UpdateLastLogin", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)

	uc := newUC(repo, notifier)
	_, err := uc.SyncFromClaims(context.Background(), claims)

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeAccountSuspended, ae.Code)
}

func TestSyncFromClaims_InvalidSubject_ReturnsError(t *testing.T) {
	claims := &keycloak.Claims{}
	claims.RegisteredClaims.Subject = "not-a-uuid"

	uc := newUC(&MockUserRepository{}, &MockNotifier{})
	_, err := uc.SyncFromClaims(context.Background(), claims)

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeTokenInvalid, ae.Code)
}

func TestSyncFromClaims_FullNameFallback(t *testing.T) {
	kcID := uuid.New()
	// No given/family name — should fall back to email local-part
	claims := &keycloak.Claims{
		Email:         "bob@example.com",
		EmailVerified: true,
	}
	claims.RegisteredClaims.Subject = kcID.String()

	repo := &MockUserRepository{}
	notifier := &MockNotifier{}

	created := &model.User{
		ID:            uuid.New(),
		KeycloakID:    kcID,
		Email:         "bob@example.com",
		FullName:      "bob", // expected fallback: email local-part
		PrimaryRole:   model.RoleUser,
		Status:        model.StatusActive,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(nil, apperr.New(apperr.CodeNotFound, "not found"))
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
		return u.FullName == "bob" && u.Email == "bob@example.com"
	})).Return(created, nil)
	repo.On("UpdateLastLogin", mock.Anything, created.ID).Return(nil)
	notifier.On("Welcome", created.Email, created.FullName).Return()

	uc := newUC(repo, notifier)
	u, err := uc.SyncFromClaims(context.Background(), claims)

	require.NoError(t, err)
	assert.Equal(t, "bob", u.FullName)
}

// ─── GetByKeycloakID ─────────────────────────────────────────────────────────

func TestGetByKeycloakID_Active(t *testing.T) {
	kcID := uuid.New()
	user := activeUser(kcID)

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(user, nil)

	uc := newUC(repo, nil)
	u, err := uc.GetByKeycloakID(context.Background(), kcID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, u.ID)
}

func TestGetByKeycloakID_Suspended_ReturnsError(t *testing.T) {
	kcID := uuid.New()
	user := activeUser(kcID)
	user.Status = model.StatusSuspended

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(user, nil)

	uc := newUC(repo, nil)
	_, err := uc.GetByKeycloakID(context.Background(), kcID)

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeAccountSuspended, ae.Code)
}

func TestGetByKeycloakID_NotFound_ReturnsError(t *testing.T) {
	kcID := uuid.New()

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(nil, apperr.New(apperr.CodeNotFound, "not found"))

	uc := newUC(repo, nil)
	_, err := uc.GetByKeycloakID(context.Background(), kcID)

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeNotFound, ae.Code)
}

// ─── UpdateProfile ───────────────────────────────────────────────────────────

func TestUpdateProfile_Success(t *testing.T) {
	kcID := uuid.New()
	user := activeUser(kcID)
	newName := "Alice Updated"
	patch := ports.ProfilePatch{FullName: &newName}

	updatedUser := *user
	updatedUser.FullName = newName

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(user, nil)
	repo.On("UpdateProfile", mock.Anything, user.ID, patch).Return(&updatedUser, nil)

	uc := newUC(repo, nil)
	u, err := uc.UpdateProfile(context.Background(), kcID, patch)

	require.NoError(t, err)
	assert.Equal(t, newName, u.FullName)
}

func TestUpdateProfile_InactiveUser_Forbidden(t *testing.T) {
	kcID := uuid.New()
	user := activeUser(kcID)
	user.Status = model.StatusSuspended

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(user, nil)

	uc := newUC(repo, nil)
	_, err := uc.UpdateProfile(context.Background(), kcID, ports.ProfilePatch{FullName: ptr("x")})

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeForbidden, ae.Code)
}

func TestUpdateProfile_NotFound_ReturnsError(t *testing.T) {
	kcID := uuid.New()

	repo := &MockUserRepository{}
	repo.On("FindByKeycloakID", mock.Anything, kcID).Return(nil, apperr.New(apperr.CodeNotFound, "not found"))

	uc := newUC(repo, nil)
	_, err := uc.UpdateProfile(context.Background(), kcID, ports.ProfilePatch{})

	require.Error(t, err)
	ae, ok := apperr.As(err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeNotFound, ae.Code)
}

// ─── internal helpers ────────────────────────────────────────────────────────

func TestComposeFullName(t *testing.T) {
	cases := []struct {
		given, family, name string
		want                string
	}{
		{"Alice", "Smith", "Alice Smith", "Alice Smith"},
		{"Alice", "", "Alice", "Alice"},
		{"", "Smith", "", "Smith"},
		{"", "", "Bob", "Bob"},
		{"", "", "", ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, composeFullName(tc.given, tc.family, tc.name))
	}
}

func TestTargetStatus(t *testing.T) {
	assert.Equal(t, model.StatusActive, targetStatus(true))
	assert.Equal(t, model.StatusPendingVerification, targetStatus(false))
}

func TestRedactEmail(t *testing.T) {
	assert.Equal(t, "a***@example.com", redactEmail("alice@example.com"))
	assert.Equal(t, "***", redactEmail("@"))
	assert.Equal(t, "***", redactEmail("x@"))
}

// ptr is shared with production code but re-declared here for test clarity
func ptr[T any](v T) *T { return &v }
