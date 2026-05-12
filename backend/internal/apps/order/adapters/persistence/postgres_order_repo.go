package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/order/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

type PostgresOrderRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepo(pool *pgxpool.Pool) *PostgresOrderRepo {
	return &PostgresOrderRepo{pool: pool}
}

const orderColumns = `
id, order_code, user_id, collector_id, material_id,
estimated_weight_kg::text, actual_weight_kg::text,
unit_price_at_order, estimated_payout, final_payout,
quality_grade, quality_bonus_pct, method,
payment_method, payment_status, paid_at,
status, otp_code, otp_verified_at, otp_expires_at,
address_text, latitude, longitude, notes,
created_at, updated_at, accepted_at, arrived_at, completed_at,
cancelled_at, cancelled_by, cancellation_reason`

// NextOrderCode reserves the next sequence value and formats it as ECC-XXXXX.
func (r *PostgresOrderRepo) NextOrderCode(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, `SELECT nextval('order_code_seq')`).Scan(&n); err != nil {
		return "", fmt.Errorf("nextval order_code: %w", err)
	}
	return fmt.Sprintf("ECC-%05d", n), nil
}

func (r *PostgresOrderRepo) Create(ctx context.Context, o *model.Order) (*model.Order, error) {
	const q = `
INSERT INTO orders (
  order_code, user_id, material_id,
  estimated_weight_kg, unit_price_at_order, estimated_payout,
  method, address_text, latitude, longitude, notes
) VALUES ($1, $2, $3, $4::numeric, $5, $6, $7::order_method, $8, $9, $10, $11)
RETURNING ` + orderColumns

	row := r.pool.QueryRow(ctx, q,
		o.OrderCode, o.UserID, o.MaterialID,
		o.EstimatedWeightKg, o.UnitPriceAtOrder, o.EstimatedPayout,
		string(o.Method), o.AddressText, o.Latitude, o.Longitude, o.Notes,
	)
	out, err := scanOrder(row)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	// Also insert the initial status_history row (created → received).
	if _, err := r.pool.Exec(ctx, `
INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, notes)
VALUES ($1, NULL, 'received'::order_status, $2, 'order created')`,
		out.ID, o.UserID); err != nil {
		return nil, fmt.Errorf("insert initial status_history: %w", err)
	}
	return out, nil
}

func (r *PostgresOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	return r.queryOne(ctx, `SELECT `+orderColumns+` FROM orders WHERE id = $1`, id)
}

func (r *PostgresOrderRepo) FindByCode(ctx context.Context, code string) (*model.Order, error) {
	return r.queryOne(ctx, `SELECT `+orderColumns+` FROM orders WHERE order_code = $1`, code)
}

func (r *PostgresOrderRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Order, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	const countQ = `SELECT count(*) FROM orders WHERE user_id = $1`
	const listQ = `SELECT ` + orderColumns + ` FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return r.queryList(ctx, countQ, listQ, []any{userID}, page, pageSize)
}

func (r *PostgresOrderRepo) ListByCollector(ctx context.Context, collectorID uuid.UUID, page, pageSize int) ([]model.Order, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	const countQ = `SELECT count(*) FROM orders WHERE collector_id = $1`
	const listQ = `SELECT ` + orderColumns + ` FROM orders WHERE collector_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	return r.queryList(ctx, countQ, listQ, []any{collectorID}, page, pageSize)
}

func (r *PostgresOrderRepo) ListIncoming(ctx context.Context, page, pageSize int) ([]model.Order, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	const countQ = `SELECT count(*) FROM orders WHERE status = 'received' AND collector_id IS NULL`
	const listQ = `SELECT ` + orderColumns + `
FROM orders WHERE status = 'received' AND collector_id IS NULL
ORDER BY created_at ASC LIMIT $1 OFFSET $2`
	return r.queryList(ctx, countQ, listQ, nil, page, pageSize)
}

func (r *PostgresOrderRepo) ListAdmin(ctx context.Context, statusFilter *model.Status, page, pageSize int) ([]model.Order, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	args := []any{}
	where := ""
	if statusFilter != nil {
		where = "WHERE status = $1::order_status"
		args = append(args, string(*statusFilter))
	}
	countQ := fmt.Sprintf(`SELECT count(*) FROM orders %s`, where)
	listQ := fmt.Sprintf(`SELECT %s FROM orders %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		orderColumns, where, len(args)+1, len(args)+2)
	return r.queryList(ctx, countQ, listQ, args, page, pageSize)
}

// Transition is the atomic state-change operation. See port doc for contract.
func (r *PostgresOrderRepo) Transition(ctx context.Context, orderID uuid.UUID, p ports.TransitionParams) (*model.Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Lock + verify current status (optimistic CAS).
	var currentStatus model.Status
	if err := tx.QueryRow(ctx,
		`SELECT status FROM orders WHERE id = $1 FOR UPDATE`,
		orderID).Scan(&currentStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "order not found")
		}
		return nil, fmt.Errorf("lock order: %w", err)
	}
	if currentStatus != p.ExpectedStatus {
		return nil, apperr.New(apperr.CodeConflict,
			fmt.Sprintf("order is in state %q, expected %q", currentStatus, p.ExpectedStatus))
	}

	// 2. Build UPDATE dynamically based on which patch fields are set.
	sets := []string{"status = $1::order_status"}
	args := []any{string(p.NewStatus)}
	i := 2

	if p.CollectorID != nil {
		sets = append(sets, fmt.Sprintf("collector_id = $%d", i))
		args = append(args, *p.CollectorID)
		i++
	}
	if p.ActualWeightKg != nil {
		sets = append(sets, fmt.Sprintf("actual_weight_kg = $%d::numeric", i))
		args = append(args, *p.ActualWeightKg)
		i++
	}
	if p.FinalPayout != nil {
		sets = append(sets, fmt.Sprintf("final_payout = $%d", i))
		args = append(args, *p.FinalPayout)
		i++
	}
	if p.QualityGrade != nil {
		sets = append(sets, fmt.Sprintf("quality_grade = $%d", i))
		args = append(args, *p.QualityGrade)
		i++
	}
	if p.QualityBonusPct != nil {
		sets = append(sets, fmt.Sprintf("quality_bonus_pct = $%d", i))
		args = append(args, *p.QualityBonusPct)
		i++
	}
	if p.OTPCode != nil {
		if *p.OTPCode == "" {
			sets = append(sets, "otp_code = NULL")
		} else {
			sets = append(sets, fmt.Sprintf("otp_code = $%d", i))
			args = append(args, *p.OTPCode)
			i++
		}
	}
	if p.OTPVerifiedAt != nil && *p.OTPVerifiedAt {
		sets = append(sets, "otp_verified_at = NOW()")
	}
	if p.OTPExpiresAt != nil {
		// Format: "clear" or "set:30m" (Go duration parsed in caller, repo trusts string)
		if *p.OTPExpiresAt == "clear" {
			sets = append(sets, "otp_expires_at = NULL")
		} else if strings.HasPrefix(*p.OTPExpiresAt, "set:") {
			d := strings.TrimPrefix(*p.OTPExpiresAt, "set:")
			sets = append(sets, fmt.Sprintf("otp_expires_at = NOW() + $%d::interval", i))
			args = append(args, d)
			i++
		}
	}
	if p.PaymentStatus != nil {
		sets = append(sets, fmt.Sprintf("payment_status = $%d", i))
		args = append(args, *p.PaymentStatus)
		i++
	}
	if p.PaidAt != nil && *p.PaidAt {
		sets = append(sets, "paid_at = NOW()")
	}
	if p.CancellationReason != nil {
		sets = append(sets, fmt.Sprintf("cancellation_reason = $%d", i))
		args = append(args, *p.CancellationReason)
		i++
		sets = append(sets, fmt.Sprintf("cancelled_by = $%d", i))
		args = append(args, p.ChangedBy)
		i++
	}
	if p.StampAcceptedAt {
		sets = append(sets, "accepted_at = NOW()")
	}
	if p.StampArrivedAt {
		sets = append(sets, "arrived_at = NOW()")
	}
	if p.StampCompletedAt {
		sets = append(sets, "completed_at = NOW()")
	}
	if p.StampCancelledAt {
		sets = append(sets, "cancelled_at = NOW()")
	}

	args = append(args, orderID)
	updateQ := fmt.Sprintf(`UPDATE orders SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(sets, ", "), i, orderColumns)

	updated, err := scanOrderTx(tx, ctx, updateQ, args...)
	if err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	// 3. Audit row.
	if _, err := tx.Exec(ctx, `
INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, notes)
VALUES ($1, $2::order_status, $3::order_status, $4, $5)`,
		orderID, string(p.ExpectedStatus), string(p.NewStatus), p.ChangedBy, p.Notes,
	); err != nil {
		return nil, fmt.Errorf("insert status_history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return updated, nil
}

func (r *PostgresOrderRepo) StatusHistory(ctx context.Context, orderID uuid.UUID) ([]model.StatusHistory, error) {
	const q = `SELECT id, order_id, from_status, to_status, changed_by, notes, changed_at
FROM order_status_history WHERE order_id = $1 ORDER BY changed_at ASC`
	rows, err := r.pool.Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("query status history: %w", err)
	}
	defer rows.Close()

	out := []model.StatusHistory{}
	for rows.Next() {
		var h model.StatusHistory
		if err := rows.Scan(&h.ID, &h.OrderID, &h.FromStatus, &h.ToStatus, &h.ChangedBy, &h.Notes, &h.ChangedAt); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// ─── helpers ─────────────────────────────────────────────────────────────

func (r *PostgresOrderRepo) queryOne(ctx context.Context, q string, args ...any) (*model.Order, error) {
	row := r.pool.QueryRow(ctx, q, args...)
	o, err := scanOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.New(apperr.CodeNotFound, "order not found")
		}
		return nil, fmt.Errorf("query order: %w", err)
	}
	return o, nil
}

func (r *PostgresOrderRepo) queryList(ctx context.Context, countQ, listQ string, baseArgs []any, page, pageSize int) ([]model.Order, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, baseArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}
	args := append(baseArgs, pageSize, (page-1)*pageSize)
	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	out := make([]model.Order, 0, pageSize)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *o)
	}
	return out, total, rows.Err()
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func scanOrder(row pgx.Row) (*model.Order, error) {
	var o model.Order
	err := row.Scan(
		&o.ID, &o.OrderCode, &o.UserID, &o.CollectorID, &o.MaterialID,
		&o.EstimatedWeightKg, &o.ActualWeightKg,
		&o.UnitPriceAtOrder, &o.EstimatedPayout, &o.FinalPayout,
		&o.QualityGrade, &o.QualityBonusPct, &o.Method,
		&o.PaymentMethod, &o.PaymentStatus, &o.PaidAt,
		&o.Status, &o.OTPCode, &o.OTPVerifiedAt, &o.OTPExpiresAt,
		&o.AddressText, &o.Latitude, &o.Longitude, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt, &o.AcceptedAt, &o.ArrivedAt, &o.CompletedAt,
		&o.CancelledAt, &o.CancelledBy, &o.CancellationReason,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// scanOrderTx is the tx-bound version (tx.QueryRow doesn't take ctx separately).
func scanOrderTx(tx pgx.Tx, ctx context.Context, q string, args ...any) (*model.Order, error) {
	return scanOrder(tx.QueryRow(ctx, q, args...))
}
