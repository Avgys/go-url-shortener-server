package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/model/responses"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service/auth"
	"go-url-shortener/internal/service/shortifier"
)

const exampleLongURL = "https://example.com/long"

type noopAudit struct{}

func (noopAudit) Publish(string, string, string) {}

type stubGen struct{}

func (stubGen) GetRandomString(int) string { return "ex1" }

func newExampleHandlers() (*handler.Handlers, context.CancelFunc) {
	logger.SetDiscardOutput(true)

	ctx, cancel := context.WithCancel(context.Background())
	store := repository.NewInMemoryStore(nil)
	sf, err := shortifier.NewShortifier(
		ctx,
		stubGen{},
		store,
		noopAudit{},
		&flagvalues.NetAddress{Scheme: "http", Host: "localhost:8080"},
	)
	if err != nil {
		panic(err)
	}
	return handler.NewHandlers(sf, store, noopAudit{}), cancel
}

func exampleClaims() *auth.TokenClaims {
	return auth.NewToken(42, "alice")
}

func exampleTeardown(cancel context.CancelFunc) {
	cancel()
	logger.SetDiscardOutput(false)
}

func withAuth(req *http.Request, claims *auth.TokenClaims) *http.Request {
	return req.WithContext(claims.WithContext(req.Context()))
}

func withShortKey(req *http.Request, shortKey string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("url", shortKey)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func shortenExampleURL(h *handler.Handlers, claims *auth.TokenClaims) (status int, shortURL, shortKey string) {
	body, err := json.Marshal(requests.ShortenReq{URL: exampleLongURL})
	if err != nil {
		panic(err)
	}

	req := withAuth(httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body)), claims)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ShortenURL(rec, req)

	var resp responses.ShortenResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		panic(err)
	}

	parsed, err := url.Parse(resp.URL)
	if err != nil {
		panic(err)
	}

	return rec.Code, resp.URL, strings.TrimPrefix(parsed.Path, "/")
}

// ExampleHandlers_ShortenURL shortens a long URL via POST /api/shorten.
func ExampleHandlers_ShortenURL() {
	h, cancel := newExampleHandlers()
	defer exampleTeardown(cancel)

	status, shortURL, _ := shortenExampleURL(h, exampleClaims())
	fmt.Printf("%d %s", status, shortURL)

	// Output: 201 http://localhost:8080/ex1
}

// ExampleHandlers_Redirect resolves a short key and redirects to the original URL.
func ExampleHandlers_Redirect() {
	h, cancel := newExampleHandlers()
	defer exampleTeardown(cancel)

	claims := exampleClaims()
	_, _, shortKey := shortenExampleURL(h, claims)

	req := withShortKey(httptest.NewRequest(http.MethodGet, "/"+shortKey, nil), shortKey)
	rec := httptest.NewRecorder()
	h.Redirect(rec, req)

	fmt.Printf("%d %s", rec.Code, rec.Header().Get("Location"))

	// Output: 307 https://example.com/long
}

// ExampleHandlers_GetURLsByUserID lists short and original URL pairs for the authenticated user.
func ExampleHandlers_GetURLsByUserID() {
	h, cancel := newExampleHandlers()
	defer exampleTeardown(cancel)

	claims := exampleClaims()
	shortenExampleURL(h, claims)

	req := withAuth(httptest.NewRequest(http.MethodGet, "/api/user/urls", nil), claims)
	rec := httptest.NewRecorder()
	h.GetURLsByUserID(rec, req)

	var pairs []responses.URLPair
	if err := json.Unmarshal(rec.Body.Bytes(), &pairs); err != nil {
		panic(err)
	}

	fmt.Printf("%d %d", rec.Code, len(pairs))

	// Output: 200 1
}

// ExampleHandlers_DeleteShortURL queues deletion of short URL keys for the authenticated user.
func ExampleHandlers_DeleteShortURL() {
	h, cancel := newExampleHandlers()
	defer exampleTeardown(cancel)

	claims := exampleClaims()
	_, _, shortKey := shortenExampleURL(h, claims)

	body, err := json.Marshal([]string{shortKey})
	if err != nil {
		panic(err)
	}

	req := withAuth(httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body)), claims)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.DeleteShortURL(rec, req)

	fmt.Println(rec.Code)

	// Output: 202
}
