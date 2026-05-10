package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	DSN       string
	Container *tcpostgres.PostgresContainer
}

func NewTestDatabase(ctx context.Context) (*TestDatabase, func(), error) {
	container, err := tcpostgres.Run(
		ctx,
		"postgres:17-alpine",
		tcpostgres.WithDatabase("activist_base_test"),
		tcpostgres.WithUsername("activist"),
		tcpostgres.WithPassword("activist"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(context.Background())
		return nil, nil, fmt.Errorf("postgres connection string: %w", err)
	}

	if err := applyMigrations(ctx, dsn); err != nil {
		_ = container.Terminate(context.Background())
		return nil, nil, err
	}

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			_ = container.Terminate(context.Background())
		})
	}

	return &TestDatabase{
		DSN:       dsn,
		Container: container,
	}, cleanup, nil
}

func applyMigrations(ctx context.Context, databaseURL string) error {
	migrateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.PingContext(migrateCtx); err != nil {
		return fmt.Errorf("ping migration connection: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	_, thisFilePath, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("resolve migrations directory: runtime caller unavailable")
	}
	migrationsDir := filepath.Join(filepath.Dir(thisFilePath), "..", "..", "..", "migrations")

	if err := goose.UpContext(migrateCtx, db, migrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
