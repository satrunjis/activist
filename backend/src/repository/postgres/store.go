package postgres

import (
	"context"
	"fmt"

	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	return &Store{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

func (s *Store) Close() {
	if s == nil || s.Pool == nil {
		return
	}
	s.Pool.Close()
}
