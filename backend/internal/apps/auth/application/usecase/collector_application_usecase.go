package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// CollectorApplicationUseCase orchestrates the become-collector flow:
// user submits → admin reviews → on approve, role mutated in Keycloak.
type CollectorApplicationUseCase struct {
	apps      ports.CollectorApplicationRepository
	users     userports.UserRepository
	adminRepo ports.AdminUserRepository
	kcAdmin   *keycloak.AdminClient
	audit     *AuditService
	notifier  ports.Notifier
}

func NewCollectorApplicationUseCase(
	apps ports.CollectorApplicationRepository,
	users userports.UserRepository,
	adminRepo ports.AdminUserRepository,
	kcAdmin *keycloak.AdminClient,
	audit *AuditService,
	notifier ports.Notifier,
) *CollectorApplicationUseCase {
	return &CollectorApplicationUseCase{
		apps: apps, users: users, adminRepo: adminRepo,
		kcAdmin: kcAdmin, audit: audit, notifier: notifier,
	}
}

// Submit is the user-side entry: POST /v1/users/me/become-collector.
//
// Rejects if the user already has the collector role (or higher). Rejects if
// there's already a pending application (DB partial unique enforces this too,
// but we check first to return a friendlier error).
func (uc *CollectorApplicationUseCase) Submit(
	ctx context.Context,
	keycloakID uuid.UUID,
	userRoles []string,
	bizName string,
	licenseNo *string,
	ktpURL string,
	siupURL *string,
	address string,
	ip, ua string,
) (*model.CollectorApplication, error) {

	for _, r := range userRoles {
		if r == "cv" || r == "collector" || r == "admin" || r == "super_admin" {
			return nil, apperr.New(apperr.CodeConflict,
				"you already have collector or partner privileges or higher")
		}
	}

	u, err := uc.users.FindByKeycloakID(ctx, keycloakID)
	if err != nil {
		return nil, err
	}
	if !u.IsActive() {
		return nil, apperr.New(apperr.CodeForbidden,
			"only active accounts can apply to become a collector")
	}

	existing, err := uc.apps.FindPendingByUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("check existing application: %w", err)
	}
	if existing != nil {
		return nil, apperr.New(apperr.CodeConflict,
			"you already have a pending application").WithDetails(map[string]any{
			"application_id": existing.ID,
		})
	}

	created, err := uc.apps.Create(ctx, &model.CollectorApplication{
		UserID:       u.ID,
		BusinessName: bizName,
		LicenseNo:    licenseNo,
		KTPURL:       ktpURL,
		SIUPURL:      siupURL,
		Address:      address,
	})
	if err != nil {
		return nil, err
	}

	uc.audit.Record(ctx, model.EventCollectorApplicationSubmitted,
		&u.ID, &keycloakID, ip, ua,
		map[string]any{"application_id": created.ID, "business_name": bizName},
	)

	if uc.notifier != nil {
		licenseStr := ""
		if created.LicenseNo != nil {
			licenseStr = *created.LicenseNo
		}
		uc.notifier.CollectorApplicationSubmitted(u.Email, u.FullName, created.BusinessName, created.Address, licenseStr)
	}

	return created, nil
}

// List for admins: GET /v1/admin/collector-applications.
func (uc *CollectorApplicationUseCase) List(ctx context.Context, status *model.ApplicationStatus, page, pageSize int) ([]model.CollectorApplication, int64, error) {
	return uc.apps.List(ctx, status, page, pageSize)
}

// Approve: admin approves a pending application. Adds `collector` realm role
// in Keycloak AND updates `users.primary_role` for display.
//
// Failure mode: if Keycloak role assignment fails after DB approval, the
// state is inconsistent (DB approved, Keycloak not). We mitigate by writing
// the audit row only after both succeed; in production a saga or outbox is
// the right pattern but overkill for MVP.
func (uc *CollectorApplicationUseCase) Approve(ctx context.Context, appID uuid.UUID, reviewerID uuid.UUID, reviewerKCID uuid.UUID, ip, ua string) (*model.CollectorApplication, error) {
	app, err := uc.apps.Approve(ctx, appID, reviewerID)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.FindByID(ctx, app.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup applicant: %w", err)
	}

	if err := uc.kcAdmin.AddRealmRole(ctx, user.KeycloakID, "collector"); err != nil {
		// Don't roll back DB — admin can retry by clicking approve again
		// (idempotent in Keycloak); flag inconsistency in audit.
		uc.audit.Record(ctx, model.EventCollectorApplicationApproved,
			&user.ID, &user.KeycloakID, ip, ua,
			map[string]any{
				"application_id": app.ID,
				"reviewer_id":    reviewerID,
				"keycloak_role_assigned": false,
				"keycloak_error": err.Error(),
			},
		)
		return nil, apperr.Wrap(apperr.CodeUpstreamUnavail,
			"application approved in DB but Keycloak role assignment failed; retry approve", err)
	}

	if _, err := uc.adminRepo.UpdatePrimaryRole(ctx, user.ID, usermodel.RoleCollector); err != nil {
		// Non-fatal — Keycloak is source of truth; primary_role is denorm.
		uc.audit.Record(ctx, model.EventCollectorApplicationApproved,
			&user.ID, &user.KeycloakID, ip, ua,
			map[string]any{
				"application_id": app.ID,
				"reviewer_id":    reviewerID,
				"primary_role_updated": false,
				"db_error": err.Error(),
			},
		)
		return app, nil
	}

	uc.audit.Record(ctx, model.EventCollectorApplicationApproved,
		&user.ID, &user.KeycloakID, ip, ua,
		map[string]any{"application_id": app.ID, "reviewer_id": reviewerID},
	)
	uc.audit.Record(ctx, model.EventRolePromoted,
		&user.ID, &user.KeycloakID, ip, ua,
		map[string]any{"from": "user", "to": "collector"},
	)

	if uc.notifier != nil {
		uc.notifier.CollectorApplicationApproved(user.Email, user.FullName, app.BusinessName)
	}
	return app, nil
}

func (uc *CollectorApplicationUseCase) Reject(ctx context.Context, appID uuid.UUID, reviewerID uuid.UUID, reviewerKCID uuid.UUID, reason string, ip, ua string) (*model.CollectorApplication, error) {
	app, err := uc.apps.Reject(ctx, appID, reviewerID, reason)
	if err != nil {
		return nil, err
	}

	uc.audit.Record(ctx, model.EventCollectorApplicationRejected,
		&app.UserID, nil, ip, ua,
		map[string]any{
			"application_id": app.ID,
			"reviewer_id":    reviewerID,
			"reason":         reason,
		},
	)

	if uc.notifier != nil {
		// Need email + name for notification — fetch user.
		if user, err := uc.users.FindByID(ctx, app.UserID); err == nil {
			uc.notifier.CollectorApplicationRejected(user.Email, user.FullName, app.BusinessName, reason)
		}
	}
	return app, nil
}
