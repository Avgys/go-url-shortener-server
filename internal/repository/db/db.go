package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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

	// if err := runMigrations(pool); err != nil {
	// 	return nil, err
	// }

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

func runMigrations(pool *pgxpool.Pool) error {

	// sourceURL := "file://migrations/sql" // usually fine

	// m, err := migrate.New(sourceURL,
	// 	cfg.ConnectionString,
	// )

	sql := `CREATE TABLE IF NOT EXISTS urls (
			id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			short_url VARCHAR(255) NOT NULL,
			long_url VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now()
		);

		CREATE INDEX IF NOT EXISTS idx_long_url ON urls(short_url); `

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, sql)

	if err != nil {
		return fmt.Errorf("couldn't open migrations, %w", err)
	}

	// if err := m.Up(); err != nil && err != migrate.ErrNoChange {
	// 	return fmt.Errorf("couldn't run migrations, %w", err)
	// }

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) Close() {
	db.Pool.Close()
}
