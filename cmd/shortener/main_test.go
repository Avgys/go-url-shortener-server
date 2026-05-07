package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	"go-url-shortener/internal/service"
	"go-url-shortener/internal/shared"
	"go-url-shortener/internal/testcommon"
)

func TestRouter(t *testing.T) {
	ts := getTestRouter(t)
	defer ts.Close()

	redirectURL := "http://someurl"
	host := ts.URL
	awaitedURL, err := url.JoinPath(host, testcommon.TestStr)
	require.NoError(t, err)

	var testTable = []struct {
		name    string
		request *http.Request
		want    testcommon.ResponseWant
	}{
		{name: "Store url", request: getStoreRequest(t, host, redirectURL),
			want: testcommon.ResponseWant{StatusCode: http.StatusCreated, Body: awaitedURL}},
		{name: "Redirect url", request: getRedirectRequest(t, host, testcommon.TestStr),
			want: testcommon.ResponseWant{StatusCode: http.StatusTemporaryRedirect, Body: "", Headers: map[string]string{"Location": redirectURL}}},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.request)
			defer resp.Body.Close()

			testcommon.CheckResponseFields(t, resp, tt.want)
		})
	}
}

func getTestRouter(t *testing.T) *httptest.Server {
	t.Helper()

	cfg, err := config.GetConfig([]string{}, &zerolog.Logger{})
	require.NoError(t, err)

	store := repository.NewInMemoryStore(nil)
	strGen := &testcommon.MockStrGen{}

	shortifier := service.NewShortifier(t.Context(), strGen, store, &cfg.RedirectDomain)
	h := &handler.Handlers{Shortifier: shortifier}

	ts := httptest.NewServer(router.NewRouter(h))

	u, err := shared.GetURL(ts.URL, true)
	require.NoError(t, err)

	cfg.RedirectDomain.Host = u.Host
	cfg.RedirectDomain.Scheme = u.Scheme

	return ts
}

func testRequest(t *testing.T, ts *httptest.Server, req *http.Request) *http.Response {
	t.Helper()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := client.Do(req)
	require.NoError(t, err)

	return res
}

func getStoreRequest(t *testing.T, host string, longURL string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, host, strings.NewReader(longURL))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	return req
}

func getRedirectRequest(t *testing.T, host string, shortURL string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, host+"/"+shortURL, nil)
	require.NoError(t, err)

	return req
}
