package dbrepos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-url-shortener/internal/db"
	dbmodel "go-url-shortener/internal/model/db"
	repoerrors "go-url-shortener/internal/repository/errors"
	urlsrepository "go-url-shortener/sqlc/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
)

type DBStore struct {
	db      *db.DB
	queries *urlsrepository.Queries
	logger  *zerolog.Logger
}

const dbOpTimeout = 1 * time.Second

func NewURLRepository(ctx context.Context, dbConn *db.DB, logger *zerolog.Logger) *DBStore {
	queries := urlsrepository.New(dbConn.Pool)
	return &DBStore{db: dbConn, queries: queries, logger: logger}
}

func (s *DBStore) ResolveShortURL(ctx context.Context, shortURL string) (dbmodel.DBURL, error) {
	row, err := s.queries.GetURLByShortURL(ctx, shortURL)

	if errors.Is(err, pgx.ErrNoRows) {
		return dbmodel.DBURL{}, repoerrors.ErrNotFound
	}

	if err != nil {
		return dbmodel.DBURL{}, err
	}

	u := dbmodel.DBURL{
		ID:          int(row.ID),
		ShortURL:    row.ShortUrl,
		OriginalURL: row.LongUrl,
		CreatedAt:   row.CreatedAt.Time,
		UserID:      row.UserID.Int64,
	}
	if row.DeletedAtUtc.Valid {
		t := row.DeletedAtUtc.Time
		u.DeletedAtUTC = &t
	}
	return u, nil
}

func (s *DBStore) StoreBatch(ctx context.Context, input map[string]string, userID int64) (retryToInsert []string, alreadyStoredURL map[string]string, err error) {

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

	fullURLArray := make([]string, 0, len(input))
	shortURLArray := make([]string, 0, len(input))
	for longURL, shortURL := range input {
		fullURLArray = append(fullURLArray, longURL)
		shortURLArray = append(shortURLArray, shortURL)
	}

	qtx := s.queries.WithTx(tx)
	rows, err := qtx.StoreBatchUrls(ctxTimeout, urlsrepository.StoreBatchUrlsParams{
		LongUrls:  fullURLArray,
		ShortUrls: shortURLArray,
		UserID:    userID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}

	alreadyStoredURL = make(map[string]string, 0)
	retryToInsert = make([]string, 0)

	for _, row := range rows {
		if row.StoredShortUrl != "" {
			alreadyStoredURL[row.LongUrl] = row.StoredShortUrl
		}
		if row.Retry {
			retryToInsert = append(retryToInsert, row.LongUrl)
		}
	}

	if err := tx.Commit(ctxTimeout); err != nil {
		return nil, nil, fmt.Errorf("failed to commit: %w", err)
	}

	return
}

func (s *DBStore) GetURLsByUserID(ctx context.Context, userID int64) ([]dbmodel.DBURL, error) {

	ctxTimeout, cancel := context.WithTimeout(ctx, dbOpTimeout)
	defer cancel()

	rows, err := s.queries.GetURLsByUserID(ctxTimeout, pgtype.Int8{Int64: userID, Valid: true})

	if err != nil {
		return nil, fmt.Errorf("failed to get URLs by user ID: %w", err)
	}

	urls := make([]dbmodel.DBURL, 0, len(rows))

	for _, row := range rows {
		u := dbmodel.DBURL{
			ID:          int(row.ID),
			ShortURL:    row.ShortUrl,
			OriginalURL: row.LongUrl,
			CreatedAt:   row.CreatedAt.Time,
			UserID:      row.UserID.Int64,
		}
		if row.DeletedAtUtc.Valid {
			t := row.DeletedAtUtc.Time
			u.DeletedAtUTC = &t
		}
		urls = append(urls, u)
	}

	return urls, nil
}

func (s *DBStore) DeleteURLS(ctx context.Context, groupedByUser map[int64][]string) ([]dbmodel.DBURL, error) {

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

	urls := make([]dbmodel.DBURL, 0)

	defer rows.Close()
	for rows.Next() {
		var dbURL dbmodel.DBURL
		var deletedAt pgtype.Timestamp

		if err = rows.Scan(&dbURL.ID, &dbURL.ShortURL, &dbURL.OriginalURL, &dbURL.CreatedAt, &dbURL.UserID, &deletedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, repoerrors.ErrNotFound
			}

			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if deletedAt.Valid {
			t := deletedAt.Time
			dbURL.DeletedAtUTC = &t
		}

		urls = append(urls, dbURL)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return urls, nil
}

func (s *DBStore) TestConnection(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *DBStore) Close() error {
	return s.db.Close()
}
