package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
	"github.com/jackc/pgx/v5"
)

type DBStore struct {
	db *db.DB
}

func NewDBStore(ctx context.Context, dbConfig *db.Config) (*DBStore, error) {
	dbConnection, err := db.NewDB(ctx, dbConfig)

	if err != nil {
		return nil, err
	}

	return &DBStore{db: dbConnection}, nil
}

func (s *DBStore) StoreURL(ctx context.Context, fullURL string, shortURL string) error {

	const queryTmp = `
		INSERT INTO public.urls (short_url, long_url)
		VALUES ($1, $2)`

	_, err := s.db.Pool.Exec(ctx, queryTmp, shortURL, fullURL)

	if err != nil {
		return fmt.Errorf("failed to insert value: %w", err)
	}

	return nil
}

func (s *DBStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	const queryTmp = `
		SELECT id, short_url, long_url, created_at
		FROM public.urls 
		WHERE short_url = $1`

	row := s.db.Pool.QueryRow(ctx, queryTmp, shortURL)

	var dbVal model.DbURL

	err := row.Scan(&dbVal.Id, &dbVal.ShortURL, &dbVal.FullURL, &dbVal.UpdateAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf("failed to scan a response row: %w", err)
	}

	return dbVal.FullURL, nil
}

func (s *DBStore) TestConnection(ctx context.Context) error {
	return s.db.Ping(ctx)
}
