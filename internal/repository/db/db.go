package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

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

	if err := runMigrations(cfg); err != nil {
		return nil, err
	}

	pool, err := initPool(ctx, cfg)

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

func runMigrations(cfg *Config) error {

	rel := filepath.Join("sql")
	abs, err := filepath.Abs(rel)
	if err != nil {
		return err
	}

	sourceURL := "file://" + filepath.ToSlash(abs) // usually fine

	m, err := migrate.New(sourceURL,
		cfg.ConnectionString,
	)

	if err != nil {
		return fmt.Errorf("couldn't open migrations, %w", err)
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
