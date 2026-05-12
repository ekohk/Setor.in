package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

type PostgresCollectorAppRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresCollectorAppRepo(pool *pgxpool.Pool) *PostgresCollectorAppRepo {
	return &PostgresCollectorAppRepo{pool: pool}
}

const appColumns = `id, user_id, business_name, license_no, ktp_url, siup_url, address,
	status, rejection_reason, reviewed_by, submitted_at, reviewed_at`

func (r *PostgresCollectorAppRepo) Create(ctx context.Context, a *model.CollectorApplication) (*model.CollectorApplication, error) {
	const q = `
INSERT INTO collector_applications (user_id, business_name, license_no, ktp_url, siup_url, address)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING ` + appColumns

	row := r.pool.QueryRow(ctx, q,
		a.UserID, a.BusinessName, a.LicenseNo, a.KTPURL, a.SIUPURL, a.Address)
	out, err := scanApp(row)
	if err != nil {
		// 23505 = unique_violation — partial unique index for one pending per user.
		if isUniqueViolation(err) {
			return nil, apperr.New(apperr.CodeConflict,
				"a pending collector application already exists for this user")
		}
		return nil, fmt.Errorf("insert collector_application: %w", err)
	}
	return out, nil
}

func (r *PostgresCollectorAppRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.CollectorApplication, error) {
	q := `SELECT ` + appColumns + ` FROM collector_applications WHERE id = $1`
	out, err := scanApp(r.pool.QueryRow(ctx, q, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "collector application not found")
		}
		return nil, fmt.Errorf("query collector_application: %w", err)
	}
	return out, nil
}

func (r *PostgresCollectorAppRepo) FindPendingByUser(ctx context.Context, userID uuid.UUID) (*model.CollectorApplication, error) {
	q := `SELECT ` + appColumns + ` FROM collector_applications WHERE user_id = $1 AND status = 'pending'`
	out, err := scanApp(r.pool.QueryRow(ctx, q, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Not-found is the normal case here; let caller decide.
			return nil, nil
		}
		return nil, fmt.Errorf("query pending app: %w", err)
	}
	return out, nil
}

func (r *PostgresCollectorAppRepo) List(ctx context.Context, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.CollectorApplication, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	args := make([]any, 0, 3)
	where := ""
	if statusFilter != nil {
		args = append(args, string(*statusFilter))
		where = "WHERE status = $1::application_status"
	}

	// Count
	countQ := fmt.Sprintf(`SELECT count(*) FROM collector_applications %s`, where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count apps: %w", err)
	}

	// Page
	limOff := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	listQ := fmt.Sprintf(`SELECT %s FROM collector_applications %s ORDER BY submitted_at DESC %s`,
		appColumns, where, limOff)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list apps: %w", err)
	}
	defer rows.Close()

	out := make([]model.CollectorApplication, 0, pageSize)
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan app: %w", err)
		}
		out = append(out, *a)
	}
	return out, total, rows.Err()
}

func (r *PostgresCollectorAppRepo) Approve(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID) (*model.CollectorApplication, error) {
	const q = `
UPDATE collector_applications
SET status = 'approved', reviewed_by = $2, reviewed_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING ` + appColumns

	out, err := scanApp(r.pool.QueryRow(ctx, q, id, reviewerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeConflict,
				"application not found or not pending")
		}
		return nil, fmt.Errorf("approve app: %w", err)
	}
	return out, nil
}

func (r *PostgresCollectorAppRepo) Reject(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, reason string) (*model.CollectorApplication, error) {
	const q = `
UPDATE collector_applications
SET status = 'rejected', rejection_reason = $3, reviewed_by = $2, reviewed_at = NOW()
WHERE id = $1 AND status = 'pending'
RETURNING ` + appColumns

	out, err := scanApp(r.pool.QueryRow(ctx, q, id, reviewerID, reason))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeConflict,
				"application not found or not pending")
		}
		return nil, fmt.Errorf("reject app: %w", err)
	}
	return out, nil
}

func scanApp(row pgx.Row) (*model.CollectorApplication, error) {
	var a model.CollectorApplication
	err := row.Scan(
		&a.ID, &a.UserID, &a.BusinessName, &a.LicenseNo, &a.KTPURL, &a.SIUPURL,
		&a.Address, &a.Status, &a.RejectionReason, &a.ReviewedBy,
		&a.SubmittedAt, &a.ReviewedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}
