package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go-url-shortener/internal/model/responses"
)

func Test_handlers_GetAll(t *testing.T) {

	urls := []string{
		"http://long-url-1.com",
		"http://long-url-2.com",
	}

	r := getRouter(t, &innerStructure{
		strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "Неожиданная ошибка при создании Cookie Jar")

	client := &http.Client{Jar: jar}

	getAllURL, err := url.JoinPath(ts.URL, "api", "user", "urls")
	require.NoError(t, err)

	for _, originalURL := range urls {
		req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(originalURL))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		res, err := client.Do(req)
		require.NoError(t, err)
		res.Body.Close()

		require.Equal(t, http.StatusCreated, res.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, getAllURL, nil)
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp []responses.URLPair
	err = json.NewDecoder(res.Body).Decode(&resp)
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
