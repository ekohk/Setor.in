package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
)

// PostgresAuthEventRepo writes append-only audit rows.
type PostgresAuthEventRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresAuthEventRepo(pool *pgxpool.Pool) *PostgresAuthEventRepo {
	return &PostgresAuthEventRepo{pool: pool}
}

func (r *PostgresAuthEventRepo) Insert(ctx context.Context, e *model.AuthEvent) error {
	const q = `
INSERT INTO auth_events (user_id, keycloak_id, event_type, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4::inet, $5, $6)
`
	// pgx accepts JSON via []byte; pass nil if empty so the column stays NULL.
	var metadata any
	if len(e.Metadata) > 0 {
		metadata = e.Metadata
	}

	_, err := r.pool.Exec(ctx, q,
		e.UserID,
		e.KeycloakID,
		string(e.EventType),
		e.IPAddress,
		e.UserAgent,
		metadata,
	)
	if err != nil {
		return fmt.Errorf("insert auth_event: %w", err)
	}
	return nil
}
