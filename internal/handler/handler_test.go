package handler_test

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/stretchr/testify/assert"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type httpWant struct {
	code int
	body string
}

func GetRandomURL(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const shortHash = "short-hash"

type mockHasher struct{}

func (m *mockHasher) GetHash(input string) string {
	return shortHash
}

type innerStructure struct {
	store      *repository.Store
	hasher     shortifier.Hasher
	shortifier *shortifier.Shortifier
	handlers   *handler.Handlers
}

func Test_handlers_Redirect(t *testing.T) {

	type caseWant struct {
		headerValue string
	}

	host := "http://localhost:8080"

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		caseWant         caseWant
		want             httpWant
	}{
		{
			name: "Get redirect",
			url:  "short-url",
			defaultStructure: &innerStructure{store: &repository.Store{
				Data: map[string]string{
					"short-url": "full-url"},
				Mux: &sync.Mutex{},
			}},
			want: httpWant{
				code: http.StatusTemporaryRedirect,
			},
			caseWant: caseWant{
				headerValue: "full-url",
			},
		},
		{
			name: "Not found url",
			url:  shortHash,
			defaultStructure: &innerStructure{
				hasher: &mockHasher{},
			},
			want: httpWant{
				code: http.StatusNotFound,
				body: fmt.Sprintf("%s %s\n", shortHash, shortifier.ErrURLNotFound.Error()),
			},
		},
		{
			name: "No url param",
			url:  "",
			want: httpWant{
				code: http.StatusBadRequest,
				body: fmt.Sprintf("%s\n", errors.ErrUnsupported),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Init
			req := httptest.NewRequest(http.MethodGet, host, nil)
			req.SetPathValue("url", tt.url)
			recorder := httptest.NewRecorder()
			structure := setup(tt.defaultStructure)

			//Run
			structure.handlers.Redirect(recorder, req)
			res := recorder.Result()

			//Check
			resBody, _ := io.ReadAll(res.Body)
			res.Body.Close()

			checkResponseFields(t, res, resBody, tt.want)
			assert.Equal(t, tt.caseWant.headerValue, res.Header.Get("Location"))
		})
	}
}

func Test_handlers_ShortifyURL(t *testing.T) {

	type caseWant struct {
		headerValue string
	}

	host := "http://localhost:8080"

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		caseWant         caseWant
		want             httpWant
	}{
		{
			name: "create shorturl",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				hasher: &mockHasher{},
			},
			want: httpWant{
				code: http.StatusCreated,
				body: fmt.Sprintf("%s/%s", host, shortHash),
			},
		},
		{
			name: "url wrong format",
			url:  "/gdfgdfhs",
			want: httpWant{
				code: http.StatusBadRequest,
				body: fmt.Sprintf("%s\n", shortifier.ErrInvalidURL.Error()),
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				hasher: &mockHasher{},
				store: &repository.Store{
					Data: map[string]string{
						shortHash: "http://long-url.com"},
					Mux: &sync.Mutex{},
				}},
			want: httpWant{
				code: http.StatusOK,
				body: fmt.Sprintf("%s/%s", host, shortHash),
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: httpWant{
				code: http.StatusBadRequest,
				body: fmt.Sprintf("%s\n", handler.ErrEmptyParamBody),
			},
		},
		{
			name: "Wrong content-type",
			url:  "http://long-url.com",
			want: httpWant{
				code: http.StatusBadRequest,
				body: fmt.Sprintf("%s: %s\n", handler.ErrWrongContentType, "text"),
			},
			contentType: "text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Init
			req := httptest.NewRequest(http.MethodGet, host, strings.NewReader(tt.url))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "text/plain")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			structure := setup(tt.defaultStructure)

			//Run
			structure.handlers.ShortifyURL(recorder, req)
			res := recorder.Result()

			//Check
			resBody, _ := io.ReadAll(res.Body)
			res.Body.Close()

			checkResponseFields(t, res, resBody, tt.want)
		})
	}
}

func Test_handlers_CreateShortURLAndRead(t *testing.T) {

	type caseWant struct {
		headerValue string
	}

	host := "http://localhost:8080"

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		caseWant         caseWant
		want             httpWant
	}{
		{
			name: "create shorturl and read",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				hasher: &mockHasher{},
			},
			want: httpWant{
				code: http.StatusTemporaryRedirect,
			},
			caseWant: caseWant{
				headerValue: "http://long-url.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Init
			structure := setup(tt.defaultStructure)

			//Get short url
			req := httptest.NewRequest(http.MethodGet, host, strings.NewReader(tt.url))
			req.Header.Set("Content-Type", "text/plain")
			recorder := httptest.NewRecorder()
			structure.handlers.ShortifyURL(recorder, req)
			res := recorder.Result()
			resBody, _ := io.ReadAll(res.Body)
			res.Body.Close()

			shortURL := string(resBody)

			//Resolve short url
			req = httptest.NewRequest(http.MethodGet, shortURL, nil)
			req.SetPathValue("url", shortHash)
			recorder = httptest.NewRecorder()
			structure.handlers.Redirect(recorder, req)
			res = recorder.Result()
			resBody, _ = io.ReadAll(res.Body)
			res.Body.Close()

			checkResponseFields(t, res, resBody, tt.want)
		})
	}
}

func setup(defaultStructure *innerStructure) *innerStructure {

	var store *repository.Store
	if defaultStructure != nil && defaultStructure.store != nil {
		store = defaultStructure.store
	} else {
		store = repository.NewStore()
	}

	var hashFunc shortifier.Hasher
	if defaultStructure != nil && defaultStructure.hasher != nil {
		hashFunc = defaultStructure.hasher
	} else {
		hashFunc = hasher.NewHasher("SomeSecret")
	}

	shortifier := shortifier.NewShortifier(hashFunc, store, "http://localhost:8080")

	h := &handler.Handlers{
		Shortifier: shortifier,
	}

	return &innerStructure{
		store:      store,
		hasher:     hashFunc,
		shortifier: shortifier,
		handlers:   h,
	}
}

func checkResponseFields(t *testing.T, res *http.Response, resBody []byte, want httpWant) {

	assert.Equal(t, want.code, res.StatusCode)

	contentType := res.Header.Get("Content-Type")

	if contentType == "application/json" {
		assert.JSONEq(t, want.body, string(resBody))
	} else {
		assert.Equal(t, want.body, string(resBody))
	}
}
