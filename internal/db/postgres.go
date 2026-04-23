package db

import (
	"context"
	"fmt"
	"time"

	"github.com/crowemi-io/crowemi-trades/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func (p *Postgres) Close() {
	if p == nil || p.Pool == nil {
		return
	}
	p.Pool.Close()
}

func NewPostgres(ctx context.Context, uri string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := sqlc.New(pool)

	ret := &Postgres{
		Pool:    pool,
		Queries: queries,
	}

	return ret, nil
}
