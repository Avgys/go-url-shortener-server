package handler_test

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var testHost url.URL = url.URL{Host: "localhost:8080", Scheme: "http"}

func GetRandomURL(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type innerStructure struct {
	store      *repository.Store
	hasher     shortifier.Hasher
	shortifier *shortifier.Shortifier
	handlers   *handler.Handlers
	config     *config.Config
}

func Test_handlers_Redirect(t *testing.T) {

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		want             testcommon.ResponseWant
	}{
		{
			name: "Get redirect",
			url:  "/short-url",
			defaultStructure: &innerStructure{
				store: &repository.Store{
					Data: map[string]string{
						"short-url": "full-url"},
					Mux: &sync.Mutex{},
				},
				config: &config.Config{URL: testHost},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusTemporaryRedirect,
				Headers:    map[string]string{"Location": "full-url"},
			},
		},
		{
			name: "Not found url",
			url:  "/" + testcommon.ShortHash,
			defaultStructure: &innerStructure{
				hasher: &testcommon.MockHasher{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusNotFound,
				Body:       fmt.Sprintf("%s %s\n", testcommon.ShortHash, shortifier.ErrURLNotFound.Error()),
			},
		},
		{
			name: "No url param",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Init

			req := httptest.NewRequest(http.MethodGet, testHost.String()+tt.url, nil)

			recorder := httptest.NewRecorder()
			r := getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()

			//Check
			resBody, _ := io.ReadAll(res.Body)
			res.Body.Close()

			testcommon.CheckResponseFields(t, res, resBody, tt.want)
		})
	}
}

func Test_handlers_ShortifyURL(t *testing.T) {

	host := "http://localhost:8080"

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name: "create shorturl",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				hasher: &testcommon.MockHasher{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusCreated,
				Body:       fmt.Sprintf("%s/%s", host, testcommon.ShortHash),
			},
		},
		{
			name: "url wrong format",
			url:  "/gdfgdfhs",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf("%s\n", shortifier.ErrInvalidURL.Error()),
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				hasher: &testcommon.MockHasher{},
				store: &repository.Store{
					Data: map[string]string{
						testcommon.ShortHash: "http://long-url.com"},
					Mux: &sync.Mutex{},
				}},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%s/%s", host, testcommon.ShortHash),
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf("%s\n", handler.ErrEmptyParamBody),
			},
		},
		{
			name: "Wrong content-type",
			url:  "http://long-url.com",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusUnsupportedMediaType,
				Body:       "",
			},
			contentType: "text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Init
			req := httptest.NewRequest(http.MethodPost, host, strings.NewReader(tt.url))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "text/plain")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()

			//Check
			resBody, _ := io.ReadAll(res.Body)
			res.Body.Close()

			testcommon.CheckResponseFields(t, res, resBody, tt.want)
		})
	}
}

func Test_handlers_CreateShortURLAndRead(t *testing.T) {

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
				hasher: &testcommon.MockHasher{},
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
			r := getRouter(tt.defaultStructure)

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
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			res.Body.Close()

			testcommon.CheckResponseFields(t, res, resBody, tt.want)
		})
	}
}

func getRouter(defaultStructure *innerStructure) *chi.Mux {

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

	shortifier := shortifier.NewShortifier(hashFunc, store)

	h := &handler.Handlers{
		Shortifier: shortifier,
	}

	return router.NewRouter(h)
}
