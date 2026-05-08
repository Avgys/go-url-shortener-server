package dbrepos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-url-shortener/internal/db"
	"go-url-shortener/internal/model"
	urlsrepository "go-url-shortener/sqlc/url"

	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type DBStore struct {
	queries *urlsrepository.Queries
	logger  *zerolog.Logger
}

const dbOpTimeout = 1 * time.Second

func NewURLRepository(ctx context.Context, db *db.DB, logger *zerolog.Logger) *DBStore {
	queries := urlsrepository.New(db.Pool)
	return &DBStore{queries: queries, logger: logger}
}

func (s *DBStore) ResolveShortURL(ctx context.Context, shortURL string) (model.DBURL, error) {
	row, err := s.queries.GetURLByShortURL(ctx, shortURL)

	if err != nil {
		return model.DBURL{}, err
	}

	return row, nil
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
			WHERE u.long_url = i.long_url and deleted_at_utc is NULL
			LIMIT 1
			) AS stored_short_url,

			-- can insert only if neither key exists
			NOT EXISTS (
			SELECT 1 FROM urls u
			WHERE (u.long_url = i.long_url OR u.short_url = i.short_url) and deleted_at_utc is NULL
			) AS should_insert,

			-- short_url collision: same short_url already used for a different long_url
			EXISTS (
			SELECT 1 FROM urls u
			WHERE u.short_url = i.short_url AND u.long_url <> i.long_url and deleted_at_utc is NULL
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

func (s *DBStore) GetURLsByUserID(ctx context.Context, userID int64) ([]model.DBURL, error) {
	rows, err := s.queries.GetURLByShortURL(ctx, userID)

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	urls := make([]DBURL, 0)

	defer rows.Close()
	for rows.Next() {
		var dbURL DBURL

		if err = rows.Scan(&dbURL.ID, &dbURL.ShortURL, &dbURL.OriginalURL, &dbURL.CreatedAt, &dbURL.UserID, &dbURL.DeletedAtUTC); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}

			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, dbURL)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return urls, nil
}

func (s *DBStore) DeleteURLS(ctx context.Context, groupedByUser map[int64][]string) ([]model.DBURL, error) {

	const queryTmp = `
		update urls 
		set deleted_at_utc = $1
		where `

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(queryTmp)

	args := make([]any, 0, len(groupedByUser)*2)
	conditions := make([]string, 0, len(groupedByUser))

	args = append(args, time.Now().UTC())

	i := 0
	for k, v := range groupedByUser {
		i = i + 2

		conditions = append(conditions, fmt.Sprintf("$%d = user_id and short_url = ANY($%d::text[])", i, i+1))
		args = append(args, k, pq.Array(v))
	}

	joinedConditions := strings.Join(conditions, " OR ")

	queryBuilder.WriteString(joinedConditions)
	queryBuilder.WriteString(" RETURNING id, short_url, long_url, created_at, user_id, deleted_at_utc")

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	query := queryBuilder.String()

	rows, err := s.db.Pool.Query(ctxTimeout, query, args...)

	if err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	urls := make([]model.DBURL, 0)

	defer rows.Close()
	for rows.Next() {
		var dbURL model.DBURL

		if err = rows.Scan(&dbURL.ID, &dbURL.ShortURL, &dbURL.OriginalURL, &dbURL.CreatedAt, &dbURL.UserID, &dbURL.DeletedAtUTC); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}

			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		urls = append(urls, dbURL)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return urls, nil
}
