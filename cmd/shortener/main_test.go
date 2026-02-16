package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/Avgys/go-url-shortener-server/internal/shared"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	ts := getTestRouter()
	defer ts.Close()

	redirectURL := "http://someurl"
	host := ts.URL
	awaitedURL, _ := url.JoinPath(host, testcommon.TestStr)

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

func getTestRouter() *httptest.Server {

	cfg, _ := config.GetConfig([]string{}, &zerolog.Logger{})
	store := repository.NewInMemoryStore(nil)
	strGen := &testcommon.MockStrGen{}

	shortifier := service.NewShortifier(strGen, store, &cfg.RedirectDomain)
	h := &handler.Handlers{Shortifier: shortifier}

	ts := httptest.NewServer(router.NewRouter(h))

	u, _ := shared.GetURL(ts.URL, true)

	cfg.RedirectDomain.Host = u.Host
	cfg.RedirectDomain.Scheme = u.Scheme

	return ts
}

func testRequest(t *testing.T, ts *httptest.Server, req *http.Request) *http.Response {

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := client.Do(req)
	require.NoError(t, err)

	return res
}

func getStoreRequest(t *testing.T, host string, longURL string) *http.Request {
	req, err := http.NewRequest(http.MethodPost, host, strings.NewReader(longURL))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	return req
}

func getRedirectRequest(t *testing.T, host string, shortURL string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, host+"/"+shortURL, nil)
	require.NoError(t, err)

	return req
}
