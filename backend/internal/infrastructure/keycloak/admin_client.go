// Admin API wrapper around gocloak.
//
// Authenticates as the `setorin-backend` service account using client_credentials.
// Token is cached in-memory and refreshed when within 60s of expiry.
package keycloak

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/shared/config"
)

// AdminClient wraps gocloak with cached service-account tokens and
// helpers tailored to Setor.in's flows (lookup user, set role, etc).
type AdminClient struct {
	gc           *gocloak.GoCloak
	realm        string
	clientID     string
	clientSecret string

	mu        sync.Mutex
	token     *gocloak.JWT
	expiresAt time.Time
}

// UserInfo is the subset of Keycloak user fields we mirror to local DB.
type UserInfo struct {
	ID            uuid.UUID
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	Enabled       bool
}

// NewAdminClient connects but does NOT authenticate yet — first auth is lazy
// on first call, so server startup doesn't fail if Keycloak is briefly down.
func NewAdminClient(cfg config.KeycloakConfig) *AdminClient {
	return &AdminClient{
		gc:           gocloak.NewClient(cfg.BaseURL),
		realm:        cfg.Realm,
		clientID:     cfg.BackendClientID,
		clientSecret: cfg.BackendClientSecret,
	}
}

// HealthCheck does a one-shot login to verify credentials at startup.
// Use only if you want a fail-fast startup; otherwise lazy auth is fine.
func (a *AdminClient) HealthCheck(ctx context.Context) error {
	if _, err := a.token0(ctx); err != nil {
		return fmt.Errorf("admin client health check: %w", err)
	}
	slog.Info("keycloak admin client healthy", "realm", a.realm, "client", a.clientID)
	return nil
}

// token0 returns a valid service-account access token, refreshing if needed.
// Refreshes when within 60s of expiry to avoid race with a request mid-flight.
func (a *AdminClient) token0(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token != nil && time.Until(a.expiresAt) > 60*time.Second {
		return a.token.AccessToken, nil
	}

	tok, err := a.gc.LoginClient(ctx, a.clientID, a.clientSecret, a.realm)
	if err != nil {
		return "", fmt.Errorf("%w: client login: %v", ErrUpstream, err)
	}
	a.token = tok
	a.expiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	return tok.AccessToken, nil
}

// GetUserByID fetches a user from Keycloak by their `sub` UUID.
// Returns ErrUserNotFound if not present.
func (a *AdminClient) GetUserByID(ctx context.Context, kcID uuid.UUID) (*UserInfo, error) {
	tok, err := a.token0(ctx)
	if err != nil {
		return nil, err
	}

	u, err := a.gc.GetUserByID(ctx, tok, a.realm, kcID.String())
	if err != nil {
		var apiErr *gocloak.APIError
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: get user: %v", ErrUpstream, err)
	}

	return mapUser(u), nil
}

// AddRealmRole assigns a realm role to a user (e.g. promote to "collector").
// Idempotent: calling twice with the same role does not error.
func (a *AdminClient) AddRealmRole(ctx context.Context, kcID uuid.UUID, roleName string) error {
	tok, err := a.token0(ctx)
	if err != nil {
		return err
	}

	role, err := a.gc.GetRealmRole(ctx, tok, a.realm, roleName)
	if err != nil {
		return fmt.Errorf("%w: get role %q: %v", ErrUpstream, roleName, err)
	}

	if err := a.gc.AddRealmRoleToUser(ctx, tok, a.realm, kcID.String(),
		[]gocloak.Role{*role}); err != nil {
		return fmt.Errorf("%w: assign role %q: %v", ErrUpstream, roleName, err)
	}
	return nil
}

// RemoveRealmRole removes a realm role from a user.
func (a *AdminClient) RemoveRealmRole(ctx context.Context, kcID uuid.UUID, roleName string) error {
	tok, err := a.token0(ctx)
	if err != nil {
		return err
	}

	role, err := a.gc.GetRealmRole(ctx, tok, a.realm, roleName)
	if err != nil {
		return fmt.Errorf("%w: get role %q: %v", ErrUpstream, roleName, err)
	}

	if err := a.gc.DeleteRealmRoleFromUser(ctx, tok, a.realm, kcID.String(),
		[]gocloak.Role{*role}); err != nil {
		return fmt.Errorf("%w: remove role %q: %v", ErrUpstream, roleName, err)
	}
	return nil
}

// LogoutUser revokes all sessions for a user (called on suspend / deletion).
func (a *AdminClient) LogoutUser(ctx context.Context, kcID uuid.UUID) error {
	tok, err := a.token0(ctx)
	if err != nil {
		return err
	}
	if err := a.gc.LogoutAllSessions(ctx, tok, a.realm, kcID.String()); err != nil {
		return fmt.Errorf("%w: logout sessions: %v", ErrUpstream, err)
	}
	return nil
}

func mapUser(u *gocloak.User) *UserInfo {
	if u == nil {
		return nil
	}
	id, _ := uuid.Parse(deref(u.ID))
	return &UserInfo{
		ID:            id,
		Email:         deref(u.Email),
		EmailVerified: derefBool(u.EmailVerified),
		FirstName:     deref(u.FirstName),
		LastName:      deref(u.LastName),
		Enabled:       derefBool(u.Enabled),
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
