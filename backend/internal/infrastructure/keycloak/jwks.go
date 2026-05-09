// Package keycloak provides JWT validation against Keycloak's JWKS endpoint
// and an admin API client wrapper.
//
// JWKS is fetched once at startup, then auto-refreshed periodically. On
// `kid` cache miss, an immediate refresh is forced before failing.
package keycloak

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/setorin/setorin/backend/internal/shared/config"
)

// Validator verifies Keycloak-issued JWTs.
type Validator struct {
	jwks    keyfunc.Keyfunc
	issuer  string
	audiences []string
	skew    time.Duration
}

// Claims is the subset of standard + custom Keycloak claims we care about.
// Use AllRoles() to get realm roles (composite-expanded by Keycloak at issue time).
type Claims struct {
	jwt.RegisteredClaims

	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Name              string `json:"name"`

	// db_user_id custom claim — populated after first /v1/auth/sync.
	// Empty for fresh tokens; backend then falls back to lookup by sub.
	DBUserID string `json:"db_user_id,omitempty"`

	RealmAccess RealmAccess `json:"realm_access"`
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

// HasRole returns true if the token's realm_access.roles contains role.
// Composite role expansion is done by Keycloak when the token is issued,
// so this is a flat membership check.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole returns true if the token has at least one of the given roles.
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, want := range roles {
		if c.HasRole(want) {
			return true
		}
	}
	return false
}

// NewValidator initializes the JWKS-backed validator.
// It does an initial fetch so the server fails fast if Keycloak is unreachable.
//
// keyfunc/v3 NewDefaultCtx auto-refreshes the JWKS in the background using
// sensible defaults (refresh interval, jitter, retry on miss). For now the
// configured JWT.JWKSCacheTTL is informational only — fine-tuning the refresh
// schedule needs the lower-level NewDefault constructor.
func NewValidator(ctx context.Context, cfg config.Config) (*Validator, error) {
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.Keycloak.JWKSURL})
	if err != nil {
		return nil, fmt.Errorf("init JWKS: %w", err)
	}

	slog.Info("JWKS validator ready",
		"jwks_url", cfg.Keycloak.JWKSURL,
		"issuer", cfg.Keycloak.Issuer,
		"audiences", cfg.JWT.Audience,
		"clock_skew", cfg.JWT.ClockSkewSeconds,
	)

	return &Validator{
		jwks:      jwks,
		issuer:    cfg.Keycloak.Issuer,
		audiences: cfg.JWT.Audience,
		skew:      time.Duration(cfg.JWT.ClockSkewSeconds) * time.Second,
	}, nil
}

// Validate parses and verifies a Bearer token. Returns claims on success.
func (v *Validator) Validate(rawToken string) (*Claims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, ErrTokenMissing
	}

	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(v.skew),
		jwt.WithExpirationRequired(),
	)

	token, err := parser.ParseWithClaims(rawToken, claims, v.jwks.Keyfunc)
	if err != nil {
		switch {
		case errorsIs(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errorsIs(err, jwt.ErrTokenSignatureInvalid),
			errorsIs(err, jwt.ErrTokenMalformed),
			errorsIs(err, jwt.ErrTokenUnverifiable):
			return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
		default:
			return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
		}
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Audience validation — Keycloak puts client IDs in `aud`. We accept the
	// token if at least one of its audiences matches one of ours.
	if len(v.audiences) > 0 && !audienceMatches(claims.Audience, v.audiences) {
		return nil, fmt.Errorf("%w: audience mismatch (got %v, want one of %v)",
			ErrTokenInvalid, claims.Audience, v.audiences)
	}

	return claims, nil
}

func audienceMatches(got jwt.ClaimStrings, want []string) bool {
	for _, w := range want {
		for _, g := range got {
			if g == w {
				return true
			}
		}
	}
	return false
}

// errorsIs is a tiny helper to avoid importing errors in two places.
func errorsIs(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
