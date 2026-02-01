package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

var testHost config.NetAddress = config.NetAddress{Host: "localhost:8080", Scheme: "http"}

type innerStructure struct {
	store      service.Repository
	strGen     service.StringGenerator
	shortifier *service.Shortifier
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
				store:  repository.NewStore(map[string]string{"short-url": "full-url"}),
				config: &config.Config{AppURL: testHost},
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
				strGen: &testcommon.MockStrGen{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusNotFound,
				Body:       fmt.Sprintf("%s %s, inner error: %s", testcommon.TestStr, service.ErrNotFound, repository.ErrNotFound),
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
			defer res.Body.Close()

			//Check
			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}

func Test_handlers_ShortifyURL(t *testing.T) {

	host := "http://localhost:8080"
	awaitedStr, _ := url.JoinPath(host, testcommon.TestStr)

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
				strGen: &testcommon.MockStrGen{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusCreated,
				Body:       awaitedStr,
			},
		},
		{
			name: "url wrong format",
			url:  "/gdfgdfhs",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       service.ErrInvalidURL.Error(),
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
				store:  repository.NewStore(map[string]string{testcommon.ShortHash: "http://long-url.com"})},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusServiceUnavailable,
				Body:       http.StatusText(http.StatusServiceUnavailable),
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       handler.ErrEmptyParamBody.Error(),
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
			defer res.Body.Close()

			//Check

			testcommon.CheckResponseFields(t, res, tt.want)
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
				strGen: &testcommon.MockStrGen{},
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
			defer res.Body.Close()

			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}

func Test_handlers_ShortenURL(t *testing.T) {

	host := "http://localhost:8080"
	requestPath, _ := url.JoinPath(host, "api", "shorten")
	awaitedStr, _ := url.JoinPath(host, testcommon.TestStr)

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
				strGen: &testcommon.MockStrGen{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusCreated,
				Body:       fmt.Sprintf(`{"result": "%s"}`, awaitedStr),
			},
		},
		{
			name: "url wrong format",
			url:  "/gdfgdfhs",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       service.ErrInvalidURL.Error(),
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
				store:  repository.NewStore(map[string]string{testcommon.ShortHash: "http://long-url.com"})},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusServiceUnavailable,
				Body:       http.StatusText(http.StatusServiceUnavailable),
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
				Body:       service.ErrEmptyURL.Error(),
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

			jsonBody, err := json.Marshal(model.ShortenReq{Url: tt.url})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			//Check

			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}

func getRouter(testStructure *innerStructure) *chi.Mux {

	if testStructure == nil {
		testStructure = &innerStructure{}
	}

	if testStructure != nil && testStructure.store == nil {
		testStructure.store = repository.NewStore(nil)
	}

	if testStructure != nil && testStructure.strGen == nil {
		testStructure.strGen = &service.StrGenerator{}
	}

	if testStructure != nil && testStructure.config == nil {
		testStructure.config, _ = config.GetConfig([]string{})
	}

	shortifier := service.NewShortifier(testStructure.strGen, testStructure.store, &testStructure.config.RedirectDomain)

	h := &handler.Handlers{
		Shortifier: shortifier,
	}

	return router.NewRouter(h)
}
