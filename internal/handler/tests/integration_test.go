package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"go-url-shortener/internal/db"
	"go-url-shortener/internal/repository/dbrepos"
	"go-url-shortener/internal/testcommon"
)

func (s *HandlerSuite) Test_handlers_CreateShortURLAndReadDbStorage() {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	rootDir := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	s.Require().NoError(os.Chdir(rootDir))
	s.T().Cleanup(func() { _ = os.Chdir(wd) })

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		s.T().Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := s.T().Context()
	prePool, err := pgxpool.New(ctx, dsn)
	s.Require().NoError(err)

	s.checkEmptyDB(ctx, prePool)

	s.T().Cleanup(func() {
		_, err = prePool.Exec(
			context.Background(),
			"DO $$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END $$;",
		)
		s.Require().NoError(err)
		prePool.Close()
	})

	dbConn, err := db.NewDB(ctx, &db.Config{ConnectionString: dsn})
	s.Require().NoError(err)

	store := dbrepos.NewURLRepository(ctx, dbConn, &zerolog.Logger{})
	s.T().Cleanup(func() {
		s.Require().NoError(store.Close())
	})

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		want             testcommon.ResponseWant
	}{
		{
			name: "create shorturl and read",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
				store:  store,
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusTemporaryRedirect,
				Headers:    map[string]string{"Location": "http://long-url.com"},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			r := s.getRouter(tt.defaultStructure)

			req := httptest.NewRequest(http.MethodPost, testHost.String(), strings.NewReader(tt.url))
			req.Header.Set("Content-Type", "text/plain")
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			resBody, err := io.ReadAll(res.Body)
			_ = res.Body.Close()

			s.Require().NoError(err)
			s.Equal(http.StatusCreated, res.StatusCode)

			shortURL := string(resBody)

			req = httptest.NewRequest(http.MethodGet, shortURL, nil)
			recorder = httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res = recorder.Result()
			defer func() { _ = res.Body.Close() }()

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}

func (s *HandlerSuite) checkEmptyDB(ctx context.Context, prePool *pgxpool.Pool) {
	var tableExists bool
	err := prePool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public')").Scan(&tableExists)
	s.Require().NoError(err)
	s.False(tableExists, "Database should be empty")
}
