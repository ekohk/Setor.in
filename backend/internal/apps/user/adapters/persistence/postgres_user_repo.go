// Package persistence holds Postgres implementations of the user module's
// repository ports. Uses pgx/v5 directly with hand-written SQL for clarity.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// PostgresUserRepo is a pgx-backed UserRepository.
type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepo constructs a UserRepository backed by Postgres.
func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

// userColumns is the canonical SELECT list — keep in sync with scanUser.
const userColumns = `id, keycloak_id, email, phone, full_name, avatar_url,
	primary_role, status, email_verified, last_login_at,
	created_at, updated_at, deleted_at`

// Create inserts a new user. Caller passes a freshly-built User with at least
// keycloak_id, email, full_name set. Other fields use DB defaults.
func (r *PostgresUserRepo) Create(ctx context.Context, u *model.User) (*model.User, error) {
	const q = `
INSERT INTO users (keycloak_id, email, full_name, phone, avatar_url, primary_role, status, email_verified)
VALUES ($1, $2, $3, $4, $5, COALESCE(NULLIF($6, '')::user_role, 'user'),
        COALESCE(NULLIF($7, '')::user_status, 'pending_verification'), $8)
RETURNING ` + userColumns

	row := r.pool.QueryRow(ctx, q,
		u.KeycloakID,
		u.Email,
		u.FullName,
		u.Phone,
		u.AvatarURL,
		string(u.PrimaryRole),
		string(u.Status),
		u.EmailVerified,
	)

	out, err := scanUser(row)
	if err != nil {
		// 23505 = unique_violation. Could be email or keycloak_id.
		if isUniqueViolation(err) {
			return nil, apperr.Wrap(apperr.CodeConflict, "user with this keycloak_id or email already exists", err)
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return out, nil
}

func (r *PostgresUserRepo) FindByKeycloakID(ctx context.Context, kcID uuid.UUID) (*model.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE keycloak_id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, kcID)
	return scanOrNotFound(row)
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, id)
	return scanOrNotFound(row)
}

func (r *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE email = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, email)
	return scanOrNotFound(row)
}

// UpdateProfile builds a dynamic UPDATE based on which patch fields are non-nil.
// Trigger trg_users_updated_at handles updated_at automatically.
func (r *PostgresUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, p ports.ProfilePatch) (*model.User, error) {
	if !p.HasAny() {
		// No-op: just return current row instead of writing nothing.
		return r.FindByID(ctx, id)
	}

	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	i := 1

	if p.FullName != nil {
		sets = append(sets, fmt.Sprintf("full_name = $%d", i))
		args = append(args, *p.FullName)
		i++
	}
	if p.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone = $%d", i))
		args = append(args, *p.Phone)
		i++
	}
	if p.AvatarURL != nil {
		sets = append(sets, fmt.Sprintf("avatar_url = $%d", i))
		args = append(args, *p.AvatarURL)
		i++
	}

	args = append(args, id)
	q := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d AND deleted_at IS NULL RETURNING %s`,
		strings.Join(sets, ", "), i, userColumns)

	row := r.pool.QueryRow(ctx, q, args...)
	out, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
		if isUniqueViolation(err) {
			return nil, apperr.Wrap(apperr.CodeConflict, "phone already in use", err)
		}
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return out, nil
}

func (r *PostgresUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET last_login_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("update last_login_at: %w", err)
	}
	return nil
}

func (r *PostgresUserRepo) UpdateAuthState(ctx context.Context, id uuid.UUID, emailVerified bool, status model.Status) (*model.User, error) {
	const q = `
UPDATE users
SET email_verified = $1,
    status = $2::user_status
WHERE id = $3 AND deleted_at IS NULL
RETURNING ` + userColumns

	row := r.pool.QueryRow(ctx, q, emailVerified, string(status), id)
	out, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("update auth state: %w", err)
	}
	return out, nil
}

// scanUser reads a row in the order of userColumns into a model.User.
func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID,
		&u.KeycloakID,
		&u.Email,
		&u.Phone,
		&u.FullName,
		&u.AvatarURL,
		&u.PrimaryRole,
		&u.Status,
		&u.EmailVerified,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// scanOrNotFound is the FindBy* convenience: maps pgx.ErrNoRows → CodeNotFound.
func scanOrNotFound(row pgx.Row) (*model.User, error) {
	out, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return out, nil
}

// isUniqueViolation checks for Postgres SQLSTATE 23505.
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}
