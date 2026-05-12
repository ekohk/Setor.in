package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/ports"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// PostgresAdminUserRepo provides admin-only user queries.
// All queries respect soft-delete (deleted_at IS NULL).
type PostgresAdminUserRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresAdminUserRepo(pool *pgxpool.Pool) *PostgresAdminUserRepo {
	return &PostgresAdminUserRepo{pool: pool}
}

const userColumns = `id, keycloak_id, email, phone, full_name, avatar_url,
	primary_role, status, email_verified, last_login_at, created_at, updated_at, deleted_at`

func (r *PostgresAdminUserRepo) List(ctx context.Context, f ports.UserListFilter) ([]usermodel.User, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	var (
		args  []any
		where = []string{"deleted_at IS NULL"}
		i     = 1
	)

	if f.Role != nil {
		where = append(where, fmt.Sprintf("primary_role = $%d::user_role", i))
		args = append(args, string(*f.Role))
		i++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d::user_status", i))
		args = append(args, string(*f.Status))
		i++
	}
	if f.Query != "" {
		where = append(where, fmt.Sprintf("(email ILIKE $%d OR full_name ILIKE $%d)", i, i))
		args = append(args, "%"+f.Query+"%")
		i++
	}

	whereSQL := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countQ := fmt.Sprintf(`SELECT count(*) FROM users %s`, whereSQL)
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	listQ := fmt.Sprintf(`SELECT %s FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		userColumns, whereSQL, i, i+1)
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	out := make([]usermodel.User, 0, f.PageSize)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, *u)
	}
	return out, total, rows.Err()
}

func (r *PostgresAdminUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status usermodel.Status) (*usermodel.User, error) {
	const q = `UPDATE users SET status = $1::user_status WHERE id = $2 AND deleted_at IS NULL RETURNING ` + userColumns
	out, err := scanUser(r.pool.QueryRow(ctx, q, string(status), id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("update user status: %w", err)
	}
	return out, nil
}

func (r *PostgresAdminUserRepo) UpdatePrimaryRole(ctx context.Context, id uuid.UUID, role usermodel.Role) (*usermodel.User, error) {
	const q = `UPDATE users SET primary_role = $1::user_role WHERE id = $2 AND deleted_at IS NULL RETURNING ` + userColumns
	out, err := scanUser(r.pool.QueryRow(ctx, q, string(role), id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found")
		}
		return nil, fmt.Errorf("update user role: %w", err)
	}
	return out, nil
}

func scanUser(row pgx.Row) (*usermodel.User, error) {
	var u usermodel.User
	err := row.Scan(
		&u.ID, &u.KeycloakID, &u.Email, &u.Phone, &u.FullName, &u.AvatarURL,
		&u.PrimaryRole, &u.Status, &u.EmailVerified, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
