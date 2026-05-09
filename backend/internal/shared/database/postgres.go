// Package database provides a pgx connection pool initializer.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/setorin/setorin/backend/internal/shared/config"
)

// New creates a pgx connection pool, pings the database, and returns the pool.
// Caller must Close() the pool on shutdown.
func New(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	pcfg.MaxConns = cfg.MaxConns
	pcfg.MinConns = cfg.MinConns
	pcfg.MaxConnLifetime = cfg.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	slog.Info("database connected",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.Name,
		"max_conns", cfg.MaxConns,
	)
	return pool, nil
}
