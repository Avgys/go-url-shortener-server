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

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
	"github.com/stretchr/testify/require"
)

type mockStrGenSequence struct {
	values []string
	idx    int
}

func (m *mockStrGenSequence) GetRandomString(n int) string {
	if len(m.values) == 0 {
		return ""
	}
	value := m.values[m.idx%len(m.values)]
	m.idx++
	return value
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
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
				store:  repository.NewInMemoryStore(map[string]string{testcommon.ShortHash: "http://long-url.com"})},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusConflict,
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
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
			},
		},
		{
			name: "url exists",
			url:  "http://long-url.com",
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
				store:  repository.NewInMemoryStore(map[string]string{testcommon.ShortHash: "http://long-url.com"})},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusConflict,
			},
		},
		{
			name: "Empty body",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
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

			jsonBody, err := json.Marshal(model.ShortenReq{URL: tt.url})
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

func Test_handlers_ShortenBatch(t *testing.T) {

	host := "http://localhost:8080"
	requestPath, _ := url.JoinPath(host, "api", "shorten", "batch")

	shortURL1, _ := url.JoinPath(host, "short-1")
	shortURL2, _ := url.JoinPath(host, "short-2")

	tests := []struct {
		name             string
		payload          model.ShortenBatchReq
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name: "create short urls",
			payload: model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "http://long-url-1.com"},
				{CorrelationID: "2", FullURL: "http://long-url-2.com"},
			},
			defaultStructure: &innerStructure{
				strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusCreated,
				Body: fmt.Sprintf(
					`[{"correlation_id":"1","short_url":"%s","is_created":true},{"correlation_id":"2","short_url":"%s","is_created":true}]`,
					shortURL1,
					shortURL2,
				),
			},
		},
		{
			name: "url exists",
			payload: model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "http://long-url.com"},
			},
			defaultStructure: &innerStructure{
				strGen: &mockStrGenSequence{values: []string{"short-1"}},
				store:  repository.NewInMemoryStore(map[string]string{"short-1": "http://long-url.com"}),
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusCreated,
				Body: fmt.Sprintf(
					`[{"correlation_id":"1","short_url":"%s","is_created":false}]`,
					shortURL1,
				),
			},
		},
		{
			name: "url wrong format",
			payload: model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "/gdfgdfhs"},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty url",
			payload: model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: ""},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "wrong content-type",
			payload: model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "http://long-url.com"},
			},
			contentType: "text",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusUnsupportedMediaType,
				Body:       "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jsonBody, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := getRouter(tt.defaultStructure)

			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}

func Test_handlers_ShortenBatchResolveByCorrelation(t *testing.T) {

	host := "http://localhost:8080"
	requestPath, _ := url.JoinPath(host, "api", "shorten", "batch")

	source := map[string]string{
		"1": "http://long-url-1.com",
		"2": "http://long-url-2.com",
	}

	payload := model.ShortenBatchReq{
		{CorrelationID: "1", FullURL: source["1"]},
		{CorrelationID: "2", FullURL: source["2"]},
	}

	jsonBody, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	r := getRouter(&innerStructure{
		strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
	})

	r.ServeHTTP(recorder, req)
	res := recorder.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusCreated, res.StatusCode)

	var batchResp model.ShortenBatchResp
	err = json.NewDecoder(res.Body).Decode(&batchResp)
	require.NoError(t, err)

	for _, item := range batchResp {
		original, ok := source[item.CorrelationID]
		require.True(t, ok)

		parsed, err := url.Parse(item.ShortURL)
		require.NoError(t, err)
		shortKey := strings.TrimPrefix(parsed.Path, "/")

		resolveReq := httptest.NewRequest(http.MethodGet, item.ShortURL, nil)
		resolveReq.SetPathValue("url", shortKey)
		resolveRecorder := httptest.NewRecorder()
		r.ServeHTTP(resolveRecorder, resolveReq)
		resolveRes := resolveRecorder.Result()
		resolveRes.Body.Close()

		require.Equal(t, http.StatusTemporaryRedirect, resolveRes.StatusCode)
		require.Equal(t, original, resolveRes.Header.Get("Location"))
	}
}
