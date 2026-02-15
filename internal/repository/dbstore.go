package repository

import (
	"context"

	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
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

func (s *DBStore) StoreURL(ctx context.Context, url string, shortURL string) error {
	return nil
}

func (s *DBStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	return "", nil
}

func (s *DBStore) TestConnection(ctx context.Context) error {
	return s.db.Ping(ctx)
}
