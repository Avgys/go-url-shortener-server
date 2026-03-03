package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/samber/lo"
)

type DBStore struct {
	db *db.DB
}

const dbOpTimeout = 1 * time.Second

func NewDBStore(ctx context.Context, dbConfig *db.Config) (*DBStore, error) {
	dbConnection, err := db.NewDB(ctx, dbConfig)

	if err != nil {
		return nil, err
	}

	return &DBStore{db: dbConnection}, nil
}

func (s *DBStore) ResolveShortURL(ctx context.Context, shortURL string) (string, error) {
	const queryTmp = `
		SELECT id, short_url, long_url, created_at
		FROM public.urls 
		WHERE short_url = $1`

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	row := s.db.Pool.QueryRow(ctxTimeout, queryTmp, shortURL)

	var dbVal model.DBURL

	err := row.Scan(&dbVal.ID, &dbVal.ShortURL, &dbVal.OriginalURL, &dbVal.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf("failed to scan a response row: %w", err)
	}

	return dbVal.OriginalURL, nil
}

func (s *DBStore) TestConnection(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DBStore) Close() error {
	s.db.Close()
	return nil
}

func (s *DBStore) StoreBatch(ctx context.Context, input Full2ShortBatch, userID int64) (retryToInsert []string, alreadyStoredURL map[string]string, err error) {

	const queryTmp = `
		WITH input(long_url, short_url) AS (
		SELECT *
		FROM unnest($1::text[], $2::text[]) AS t(long_url, short_url)
		),		

		marked AS (
			SELECT
			i.long_url,
			i.short_url,

			-- if long_url already exists, what short_url is stored for it?
			(
			SELECT u.short_url
			FROM urls u
			WHERE u.long_url = i.long_url
			LIMIT 1
			) AS stored_short_url,

			-- can insert only if neither key exists
			NOT EXISTS (
			SELECT 1 FROM urls u
			WHERE u.long_url = i.long_url OR u.short_url = i.short_url
			) AS should_insert,

			-- short_url collision: same short_url already used for a different long_url
			EXISTS (
			SELECT 1 FROM urls u
			WHERE u.short_url = i.short_url AND u.long_url <> i.long_url
			) AS retry

			FROM input i
		),

		inserted AS (
			INSERT INTO urls (long_url, short_url, user_id)
			SELECT long_url, short_url, $3
			FROM marked
			WHERE should_insert
		)

		SELECT
		m.short_url,
		m.long_url,
		m.stored_short_url,
		m.retry
		FROM marked m
		WHERE retry OR stored_short_url IS NOT NULL;`

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	tx, err := s.db.Pool.Begin(ctxTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctxTimeout)
		}
	}()

	fullURLArray := lo.Map(lo.Keys(input), func(x string, _ int) string { return x })
	shortURLArray := lo.Map(lo.Values(input), func(x string, _ int) string { return x })

	rows, err := tx.Query(ctxTimeout, queryTmp, pq.Array(fullURLArray), pq.Array(shortURLArray), userID)

	if err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}

	alreadyStoredURL = make(map[string]string, 0)
	retryToInsert = make([]string, 0)

	defer rows.Close()
	for rows.Next() {
		var shortURL, longURL, storedShortURL string
		var retry bool

		if err := rows.Scan(&shortURL, &longURL, &storedShortURL, &retry); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if storedShortURL != "" {
			alreadyStoredURL[longURL] = storedShortURL
		}

		if retry {
			retryToInsert = append(retryToInsert, longURL)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return nil, nil, fmt.Errorf("failed to commit: %w", err)
	}

	return
}

func (s *DBStore) GetURLsByUserID(ctx context.Context, userID int64) ([]*model.DBURL, error) {
	const queryTmp = `
		SELECT short_url, long_url
		FROM public.urls 
		WHERE user_id = $1`

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	rows, err := s.db.Pool.Query(ctxTimeout, queryTmp, userID)

	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	urls := make([]*model.DBURL, 0)
	for rows.Next() {
		var dbURL model.DBURL

		if err = rows.Scan(&dbURL.ShortURL, &dbURL.OriginalURL); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}

			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, &dbURL)
	}

	result := lo.Map(urls, func(dbPair *model.DBURL, _ int) *model.DBURL {
		return &model.DBURL{OriginalURL: dbPair.OriginalURL, ShortURL: dbPair.ShortURL}
	})

	return result, nil
}
