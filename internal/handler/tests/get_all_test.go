package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/auth"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/stretchr/testify/require"
)

func Test_handlers_GetAll(t *testing.T) {
	host := "http://localhost:8080"
	shortifyURL := host
	getAllURL, _ := url.JoinPath(host, "api", "user", "urls")

	urls := []string{
		"http://long-url-1.com",
		"http://long-url-2.com",
	}

	r := getRouter(&innerStructure{
		strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
	})

	var authCookie *http.Cookie

	for _, originalURL := range urls {
		req := httptest.NewRequest(http.MethodPost, shortifyURL, strings.NewReader(originalURL))
		req.Header.Set("Content-Type", "text/plain")
		if authCookie != nil {
			req.AddCookie(authCookie)
		}

		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		res := recorder.Result()
		res.Body.Close()

		require.Equal(t, http.StatusCreated, res.StatusCode)
		authCookie = requireAuthCookie(t, res)
	}

	req := httptest.NewRequest(http.MethodGet, getAllURL, nil)
	req.AddCookie(authCookie)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	res := recorder.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp []model.URLPair
	err := json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp, len(urls))

	originals := map[string]bool{}
	for _, u := range urls {
		originals[u] = true
	}

	for _, item := range resp {
		require.True(t, originals[item.OriginalURL])
		parsed, err := url.Parse(item.ShortURL)
		require.NoError(t, err)
		require.NotEmpty(t, strings.TrimPrefix(parsed.Path, "/"))
	}
}

func requireAuthCookie(t *testing.T, res *http.Response) *http.Cookie {
	for _, cookie := range res.Cookies() {
		if cookie.Name == auth.AuthCookie {
			return cookie
		}
	}

	require.FailNow(t, "auth cookie is not set")
	return nil
}
