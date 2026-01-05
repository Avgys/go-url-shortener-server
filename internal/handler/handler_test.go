package handler

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type httpWant struct {
	code int
	body string
}

func GetRandomUrl(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type innerStructure struct {
	store      *repository.Store
	hasher     *hasher.Hasher
	shortifier *shortifier.Shortifier
	handlers   *handlers
}

func Test_handlers_Redirect(t *testing.T) {

	type caseWant struct {
		headerValue string
	}

	tests := []struct {
		name         string
		url          string
		defaultStore *repository.Store
		caseWant     caseWant
		want         httpWant
	}{
		{
			name: "Get redirect",
			url:  "short-url",
			defaultStore: &repository.Store{
				Data: map[string]string{
					"short-url": "full-url"},
				Mux: &sync.Mutex{},
			},
			want: httpWant{
				code: http.StatusTemporaryRedirect,
			},
			caseWant: caseWant{
				headerValue: "full-url",
			},
		},
		{
			name: "Not found url",
			url:  GetRandomUrl(6),
			want: httpWant{
				code: http.StatusNotFound,
				body: fmt.Sprintf("%s\n", shortifier.ErrUrlNotFound.Error()),
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%v", tt.url), nil)
			req.SetPathValue("url", tt.url)
			recorder := httptest.NewRecorder()
			structure := setup(tt.defaultStore)

			//Run
			structure.handlers.Redirect(recorder, req)
			res := recorder.Result()

			//Check
			checkResponseFields(t, res, tt.want)
			assert.Equal(t, tt.caseWant.headerValue, res.Header.Get("Location"))
		})
	}
}

func Test_handlers_ShortifyURL(t *testing.T) {

	type caseWant struct {
		headerValue string
	}

	tests := []struct {
		name         string
		url          string
		defaultStore *repository.Store
		caseWant     caseWant
		want         httpWant
	}{
		{
			name: "Get redirect",
			url:  "short-url",
			defaultStore: &repository.Store{
				Data: map[string]string{
					"short-url": "full-url"},
				Mux: &sync.Mutex{},
			},
			want: httpWant{
				code: http.StatusTemporaryRedirect,
			},
			caseWant: caseWant{
				headerValue: "full-url",
			},
		},
		{
			name: "Not found url",
			url:  GetRandomUrl(6),
			want: httpWant{
				code: http.StatusNotFound,
				body: fmt.Sprintf("%s\n", shortifier.ErrUrlNotFound.Error()),
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
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%v", tt.url), nil)
			req.SetPathValue("url", tt.url)
			recorder := httptest.NewRecorder()
			structure := setup(tt.defaultStore)

			//Run
			structure.handlers.Redirect(recorder, req)
			res := recorder.Result()

			//Check
			checkResponseFields(t, res, tt.want)
			assert.Equal(t, tt.caseWant.headerValue, res.Header.Get("Location"))
		})
	}
}

func setup(defaultStore *repository.Store) *innerStructure {

	var store *repository.Store
	if defaultStore != nil {
		store = defaultStore
	} else {
		store = repository.NewStore()
	}

	hashFunc := hasher.NewHasher("SomeSecret")
	shortifier := shortifier.NewShortifier(hashFunc, store, "http://localhost:8080")

	h := &handlers{
		shortifier: shortifier,
	}

	return &innerStructure{
		store:      store,
		hasher:     hashFunc,
		shortifier: shortifier,
		handlers:   h,
	}
}

func checkResponseFields(t *testing.T, res *http.Response, want httpWant) {

	assert.Equal(t, want.code, res.StatusCode)
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)

	require.NoError(t, err)

	contentType := res.Header.Get("Content-Type")

	if contentType == "application/json" {
		assert.JSONEq(t, want.body, string(resBody))
	} else {
		assert.Equal(t, want.body, string(resBody))
	}
}
