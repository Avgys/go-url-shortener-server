package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	ConnectionString string
}

type DB struct {
	Pool *pgxpool.Pool
}

const (
	initTimeout = 30 * time.Second
)

func NewDB(ctx context.Context, cfg *Config) (*DB, error) {

	if cfg.ConnectionString == "" {
		return nil, errors.New("empty connection string")
	}

	initCtx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	pool, err := initPool(initCtx, cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err := runMigrations(initCtx, cfg); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// go func() {
	// 	<-ctx.Done()

	// 	pool.Close()
	// }()

	return &DB{Pool: pool}, nil
}

func initPool(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnectionString)

	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}

	return pool, nil
}

func runMigrations(ctx context.Context, cfg *Config) error {

	m, err := migrate.New("file://migrations/sql",
		cfg.ConnectionString,
	)

	if err != nil {
		dir, _ := os.Getwd()
		return fmt.Errorf("couldn't open migrations, %w, current dir %s", err, dir)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("couldn't run migrations, %w", err)
	}

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) Close() error {
	db.Pool.Close()
	return nil
}
