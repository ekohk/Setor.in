// Package usecase orchestrates business logic for the user module.
//
// Use cases compose ports + domain helpers; they do NOT touch HTTP or DB
// driver code directly. Errors returned should be *AppError so handlers can
// translate them via response.Err().
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// UserUseCase is the application service for user-related operations.
type UserUseCase struct {
	repo     ports.UserRepository
	kcAdmin  *keycloak.AdminClient
	notifier ports.Notifier
}

func NewUserUseCase(repo ports.UserRepository, kcAdmin *keycloak.AdminClient, notifier ports.Notifier) *UserUseCase {
	return &UserUseCase{repo: repo, kcAdmin: kcAdmin, notifier: notifier}
}

// SyncFromClaims is the entry point for /v1/auth/sync.
//
// Idempotent: creates a new local user if none exists for `claims.sub`,
// otherwise refreshes email_verified + status + last_login_at to keep DB
// in sync with Keycloak.
//
// Why both create AND update here: this endpoint is called by the frontend
// after EVERY successful login, so it must handle "first login after register"
// (create) and "subsequent login" (update) in one shot.
func (uc *UserUseCase) SyncFromClaims(ctx context.Context, claims *keycloak.Claims) (*model.User, error) {
	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeTokenInvalid, "subject is not a valid uuid", err)
	}

	existing, err := uc.repo.FindByKeycloakID(ctx, kcID)
	if err != nil && !isNotFound(err) {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	// First sync ever — create the local row.
	if existing == nil {
		fullName := composeFullName(claims.GivenName, claims.FamilyName, claims.Name)
		if fullName == "" {
			// Fall back to email local-part so DB constraint (NOT NULL) holds.
			fullName = strings.SplitN(claims.Email, "@", 2)[0]
		}

		newUser := &model.User{
			KeycloakID:    kcID,
			Email:         claims.Email,
			FullName:      fullName,
			PrimaryRole:   model.RoleUser,
			Status:        targetStatus(claims.EmailVerified),
			EmailVerified: claims.EmailVerified,
		}

		created, err := uc.repo.Create(ctx, newUser)
		if err != nil {
			return nil, err
		}
		if err := uc.repo.UpdateLastLogin(ctx, created.ID); err != nil {
			// Non-fatal — just log; the user is still synced.
			slog.Warn("update last_login failed after create", "user_id", created.ID, "err", err)
		}
		slog.Info("user synced (created)",
			"user_id", created.ID,
			"keycloak_id", kcID,
			"email", redactEmail(created.Email),
		)

		// Welcome email is fire-and-forget; failures logged inside the notifier.
		if uc.notifier != nil && created.EmailVerified {
			uc.notifier.Welcome(created.Email, created.FullName)
		}
		return created, nil
	}

	// Existing user — refresh state from claims.
	if existing.EmailVerified != claims.EmailVerified || existing.Status != targetStatus(claims.EmailVerified) {
		updated, err := uc.repo.UpdateAuthState(ctx, existing.ID, claims.EmailVerified, targetStatus(claims.EmailVerified))
		if err != nil {
			return nil, err
		}
		existing = updated
	}

	if err := uc.repo.UpdateLastLogin(ctx, existing.ID); err != nil {
		slog.Warn("update last_login failed", "user_id", existing.ID, "err", err)
	}

	if existing.Status == model.StatusSuspended {
		return nil, apperr.New(apperr.CodeAccountSuspended, "your account has been suspended; contact support")
	}
	if existing.IsDeleted() {
		return nil, apperr.New(apperr.CodeAccountDeleted, "this account no longer exists")
	}

	slog.Info("user synced (existing)",
		"user_id", existing.ID,
		"keycloak_id", kcID,
		"email", redactEmail(existing.Email),
	)
	return existing, nil
}

// GetByKeycloakID is used by /v1/auth/me and /v1/users/me.
// Returns CodeNotFound if local row missing — frontend should call /v1/auth/sync first.
func (uc *UserUseCase) GetByKeycloakID(ctx context.Context, kcID uuid.UUID) (*model.User, error) {
	u, err := uc.repo.FindByKeycloakID(ctx, kcID)
	if err != nil {
		return nil, err
	}
	if u.Status == model.StatusSuspended {
		return nil, apperr.New(apperr.CodeAccountSuspended, "your account has been suspended")
	}
	return u, nil
}

// UpdateProfile is the application logic for PATCH /v1/users/me.
func (uc *UserUseCase) UpdateProfile(ctx context.Context, kcID uuid.UUID, p ports.ProfilePatch) (*model.User, error) {
	u, err := uc.repo.FindByKeycloakID(ctx, kcID)
	if err != nil {
		return nil, err
	}
	if !u.IsActive() {
		return nil, apperr.New(apperr.CodeForbidden, "cannot update profile of inactive account")
	}
	return uc.repo.UpdateProfile(ctx, u.ID, p)
}

// composeFullName picks the best display name from claims.
// Priority: given+family > name > "".
func composeFullName(given, family, name string) string {
	g, f := strings.TrimSpace(given), strings.TrimSpace(family)
	if g != "" || f != "" {
		return strings.TrimSpace(g + " " + f)
	}
	return strings.TrimSpace(name)
}

// targetStatus returns the user_status that matches the verified flag.
// Email verified → active; not verified → pending_verification.
func targetStatus(verified bool) model.Status {
	if verified {
		return model.StatusActive
	}
	return model.StatusPendingVerification
}

func isNotFound(err error) bool {
	ae, ok := apperr.As(err)
	return ok && ae.Code == apperr.CodeNotFound
}

// redactEmail keeps logs PII-safe: alice@example.com → a***@example.com.
func redactEmail(e string) string {
	at := strings.IndexByte(e, '@')
	if at <= 1 {
		return "***"
	}
	return e[:1] + "***" + e[at:]
}

// Sentinel for callers that want to react specifically to "not found".
var ErrNotFound = errors.New("user: not found")
