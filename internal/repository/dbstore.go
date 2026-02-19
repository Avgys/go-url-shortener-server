package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/samber/lo"
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

	var dbVal model.DBURL

	err := row.Scan(&dbVal.ID, &dbVal.ShortURL, &dbVal.FullURL, &dbVal.UpdateAt)

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

func (s *DBStore) Close() error {
	s.db.Close()
	return nil
}

// marked AS (
// SELECT
// 	i.long_url,
// 	i.short_url,
// 	o.short_url as stored_short_url,
// 	(o.short_url IS NULL AND o.long_url IS NULL) AS should_insert,
// 	(o.short_url = i.short_url AND o.long_url != i.long_url) AS retry,
// 	(o.short_url != i.short_url AND o.long_url = i.long_url OR
// 		o.short_url = i.short_url AND o.long_url = i.long_url) AS already_exists
// FROM input i
// LEFT JOIN urls o
// 	ON o.long_url = i.long_url OR o.short_url = i.short_url
// ),

func (s *DBStore) StoreBatch(ctx context.Context, input Full2ShortBatch) (retryToInsert []string, alreadyStoredURL map[string]string, err error) {

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
			INSERT INTO urls (long_url, short_url)
			SELECT long_url, short_url
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

	tx, _ := s.db.Pool.Begin(ctx)

	fullURLArray := lo.Map(lo.Keys(input), func(x string, _ int) string { return x })
	shortURLArray := lo.Map(lo.Values(input), func(x string, _ int) string { return x })

	st, err := tx.Prepare(ctx, "insert", queryTmp)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare: %w", err)
	}

	rows, err := tx.Query(ctx, st.Name, pq.Array(fullURLArray), pq.Array(shortURLArray))

	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		return nil, nil, fmt.Errorf("rows error: %w, rollback error: %w", err, rollbackErr)
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
		rollbackErr := tx.Rollback(ctx)
		return nil, nil, fmt.Errorf("rows error: %w, rollback error: %w", err, rollbackErr)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit: %w", err)
	}

	return
}
