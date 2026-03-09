package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Avgys/go-url-shortener-server/internal/repository"
	repositorydb "github.com/Avgys/go-url-shortener-server/internal/repository/db"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/stretchr/testify/require"
)

func Test_handlers_CreateShortURLAndReadDbStorage(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	rootDir := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	require.NoError(t, os.Chdir(rootDir))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	dsn := "postgres://app:secret@localhost:5432/test?sslmode=disable" //os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set")
	}

	ctx := t.Context()
	prePool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	checkEmptyDb(ctx, t, err, prePool)

	t.Cleanup(func() {
		_, err = prePool.Exec(
			context.Background(),
			"DO $$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END $$;",
		)
		require.NoError(t, err)
		prePool.Close()
	})

	store, err := repository.NewDBStore(ctx, &repositorydb.Config{ConnectionString: dsn}, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = store.Close()
		require.NoError(t, err)
	})

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
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
		t.Run(tt.name, func(t *testing.T) {

			//Init
			r := getRouter(t, tt.defaultStructure)

			//Get short url
			req := httptest.NewRequest(http.MethodPost, testHost.Host, strings.NewReader(tt.url))
			req.Header.Set("Content-Type", "text/plain")
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			resBody, err := io.ReadAll(res.Body)
			res.Body.Close()

			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, res.StatusCode)

			shortURL := string(resBody)

			//Resolve short url
			req = httptest.NewRequest(http.MethodGet, shortURL, nil)
			req.SetPathValue("url", testcommon.ShortHash)
			recorder = httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res = recorder.Result()
			defer res.Body.Close()

			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}

func checkEmptyDb(ctx context.Context, t *testing.T, err error, prePool *pgxpool.Pool) {
	t.Helper()

	var tableExists bool
	err = prePool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'urls')").Scan(&tableExists)
	require.NoError(t, err)
	require.False(t, tableExists, "Database should be empty")
}
