package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {

	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
