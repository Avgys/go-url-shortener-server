package db

import (
	"context"
	"errors"
	"fmt"
	"os"

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

func NewDB(ctx context.Context, cfg *Config) (*DB, error) {

	if cfg.ConnectionString == "" {
		return nil, errors.New("empty connection string")
	}

	pool, err := initPool(ctx, cfg)

	if err := runMigrations(ctx, cfg); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func initPool(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnectionString)

	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

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

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("couldn't run migrations, %w", err)
	}

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) Close() {
	db.Pool.Close()
}
