package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// AdminUserUseCase covers admin operations on users:
// list, suspend/unsuspend, change primary role.
type AdminUserUseCase struct {
	users    ports.AdminUserRepository
	kcAdmin  *keycloak.AdminClient
	audit    *AuditService
	notifier ports.Notifier
}

func NewAdminUserUseCase(users ports.AdminUserRepository, kcAdmin *keycloak.AdminClient, audit *AuditService, notifier ports.Notifier) *AdminUserUseCase {
	return &AdminUserUseCase{users: users, kcAdmin: kcAdmin, audit: audit, notifier: notifier}
}

func (uc *AdminUserUseCase) List(ctx context.Context, f ports.UserListFilter) ([]usermodel.User, int64, error) {
	return uc.users.List(ctx, f)
}

// UpdateStatus suspends or unsuspends a user. On suspend, also revokes their
// Keycloak sessions so the next request fails with 401.
func (uc *AdminUserUseCase) UpdateStatus(
	ctx context.Context,
	targetID uuid.UUID,
	newStatus usermodel.Status,
	reason *string,
	actorID uuid.UUID,
	ip, ua string,
) (*usermodel.User, error) {
	if newStatus != usermodel.StatusActive && newStatus != usermodel.StatusSuspended {
		return nil, apperr.New(apperr.CodeValidation,
			"status must be 'active' or 'suspended'")
	}

	updated, err := uc.users.UpdateStatus(ctx, targetID, newStatus)
	if err != nil {
		return nil, err
	}

	if newStatus == usermodel.StatusSuspended {
		if err := uc.kcAdmin.LogoutUser(ctx, updated.KeycloakID); err != nil {
			// Log but don't fail — status is already suspended in DB; backend
			// auth middleware will reject on next sync via CodeAccountSuspended.
			uc.audit.Record(ctx, model.EventAccountSuspended,
				&updated.ID, &updated.KeycloakID, ip, ua,
				map[string]any{
					"reason":            valueOr(reason, ""),
					"actor_id":          actorID,
					"keycloak_logout_failed": err.Error(),
				},
			)
			if uc.notifier != nil {
				uc.notifier.AccountSuspended(updated.Email, updated.FullName, valueOr(reason, ""))
			}
			return updated, nil
		}
		uc.audit.Record(ctx, model.EventAccountSuspended,
			&updated.ID, &updated.KeycloakID, ip, ua,
			map[string]any{"reason": valueOr(reason, ""), "actor_id": actorID},
		)
		if uc.notifier != nil {
			uc.notifier.AccountSuspended(updated.Email, updated.FullName, valueOr(reason, ""))
		}
	} else {
		uc.audit.Record(ctx, model.EventAccountUnsuspended,
			&updated.ID, &updated.KeycloakID, ip, ua,
			map[string]any{"reason": valueOr(reason, ""), "actor_id": actorID},
		)
		if uc.notifier != nil {
			uc.notifier.AccountUnsuspended(updated.Email, updated.FullName)
		}
	}
	return updated, nil
}

// UpdateRole is the most privileged operation: super_admin only.
// Mutates Keycloak realm roles + DB primary_role denormalization.
//
// Strategy: fetch current roles from caller's understanding (not Keycloak round-trip),
// remove ALL non-default roles, then add the target. Simpler than diff-add-remove.
//
// For MVP we just add the target role; existing roles stay. The denorm column
// reflects the highest role.
func (uc *AdminUserUseCase) UpdateRole(
	ctx context.Context,
	targetID uuid.UUID,
	newRole usermodel.Role,
	reason *string,
	actorID uuid.UUID,
	ip, ua string,
) (*usermodel.User, error) {
	if !newRole.IsValid() {
		return nil, apperr.New(apperr.CodeValidation, "invalid role")
	}

	updated, err := uc.users.UpdatePrimaryRole(ctx, targetID, newRole)
	if err != nil {
		return nil, err
	}
	// Note: `updated.PrimaryRole` is already the NEW role (just written).
	// We don't have the old role from the repo signature; for notification
	// purposes we report the change as "to: newRole" — the email template
	// shows current vs new based on what we pass.
	oldRole := "(sebelumnya)"

	if err := uc.kcAdmin.AddRealmRole(ctx, updated.KeycloakID, string(newRole)); err != nil {
		uc.audit.Record(ctx, model.EventRoleChanged,
			&updated.ID, &updated.KeycloakID, ip, ua,
			map[string]any{
				"to_role":         string(newRole),
				"actor_id":        actorID,
				"keycloak_assign_failed": err.Error(),
			},
		)
		return nil, apperr.Wrap(apperr.CodeUpstreamUnavail,
			fmt.Sprintf("DB role updated but Keycloak role %q assign failed; retry", newRole), err)
	}

	uc.audit.Record(ctx, model.EventRoleChanged,
		&updated.ID, &updated.KeycloakID, ip, ua,
		map[string]any{
			"to_role":  string(newRole),
			"actor_id": actorID,
			"reason":   valueOr(reason, ""),
		},
	)

	if uc.notifier != nil {
		uc.notifier.RoleChanged(updated.Email, updated.FullName, oldRole, string(newRole), valueOr(reason, ""))
	}
	return updated, nil
}

func valueOr(p *string, fallback string) string {
	if p == nil {
		return fallback
	}
	return *p
}
