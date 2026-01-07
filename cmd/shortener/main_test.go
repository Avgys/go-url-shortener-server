package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	ts := getTestRouter()
	defer ts.Close()

	redirectUrl := "http://someurl"
	host := ts.URL

	var testTable = []struct {
		name    string
		request *http.Request
		want    testcommon.ResponseWant
	}{
		{name: "Store url", request: getStoreRequest(t, host, redirectUrl),
			want: testcommon.ResponseWant{StatusCode: http.StatusCreated, Body: fmt.Sprintf("%s/%s", host, testcommon.ShortHash)}},
		{name: "Redirect url", request: getRedirectRequest(t, host, testcommon.ShortHash),
			want: testcommon.ResponseWant{StatusCode: http.StatusTemporaryRedirect, Body: "", Headers: map[string]string{"Location": redirectUrl}}},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			resp, get := testRequest(t, ts, tt.request)
			testcommon.CheckResponseFields(t, resp, get, tt.want)
		})
	}
}

func getTestRouter() *httptest.Server {

	store := repository.NewStore()
	hashFunc := &testcommon.MockHasher{}

	shortifier := shortifier.NewShortifier(hashFunc, store)
	h := &handler.Handlers{Shortifier: shortifier}

	ts := httptest.NewServer(router.NewRouter(h))
	return ts
}

func testRequest(t *testing.T, ts *httptest.Server, req *http.Request) (*http.Response, []byte) {

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, resBody
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
