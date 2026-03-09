package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

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
	awaitedStr, err := url.JoinPath(host, testcommon.TestStr)
	require.NoError(t, err)

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
				store:  repository.NewInMemoryStore([]*model.DBURL{{OriginalURL: "http://long-url.com", ShortURL: testcommon.ShortHash}})},
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
			r := getRouter(t, tt.defaultStructure)

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
			r := getRouter(t, tt.defaultStructure)

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
	requestPath, err := url.JoinPath(host, "api", "shorten")
	require.NoError(t, err)

	awaitedStr, err := url.JoinPath(host, testcommon.TestStr)
	require.NoError(t, err)

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
				store:  repository.NewInMemoryStore([]*model.DBURL{&model.DBURL{OriginalURL: "http://long-url.com", ShortURL: testcommon.ShortHash}})},
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
			r := getRouter(t, tt.defaultStructure)

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
	requestPath, err := url.JoinPath(host, "api", "shorten", "batch")
	require.NoError(t, err)

	shortURL1, err := url.JoinPath(host, "short-1")
	require.NoError(t, err)

	shortURL2, err := url.JoinPath(host, "short-2")
	require.NoError(t, err)

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
				store:  repository.NewInMemoryStore([]*model.DBURL{&model.DBURL{OriginalURL: "http://long-url.com", ShortURL: "short-1"}}),
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
			r := getRouter(t, tt.defaultStructure)

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
	r := getRouter(t, &innerStructure{
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

func Test_handlers_ShortenDeleteRead(t *testing.T) {
	const workerCount = 20
	const urlsPerWorker = 3

	waitForGone := func(client *http.Client, resolveURL string, shortKey string) error {
		deadline := time.Now().Add(2 * time.Second)
		for {
			resolveReq, err := http.NewRequest(http.MethodGet, resolveURL, nil)
			if err != nil {
				return err
			}
			resolveReq.SetPathValue("url", shortKey)
			resolveRes, err := client.Do(resolveReq)
			if err != nil {
				return err
			}
			resolveRes.Body.Close()

			if resolveRes.StatusCode == http.StatusGone {
				return nil
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("expected %d, got %d", http.StatusGone, resolveRes.StatusCode)
			}

			time.Sleep(20 * time.Millisecond)
		}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			jar, err := cookiejar.New(nil)
			if err != nil {
				errCh <- err
				return
			}

			client := &http.Client{
				Jar: jar,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			shortValues := []string{
				fmt.Sprintf("s-%d-1", idx),
				fmt.Sprintf("s-%d-2", idx),
				fmt.Sprintf("s-%d-3", idx),
			}

			payload := model.ShortenBatchReq{
				{CorrelationID: "1", FullURL: fmt.Sprintf("http://long-url-%d-1.com", idx)},
				{CorrelationID: "2", FullURL: fmt.Sprintf("http://long-url-%d-2.com", idx)},
				{CorrelationID: "3", FullURL: fmt.Sprintf("http://long-url-%d-3.com", idx)},
			}

			jsonBody, err := json.Marshal(payload)
			if err != nil {
				errCh <- err
				return
			}

			r := getRouter(t, &innerStructure{
				strGen: &mockStrGenSequence{values: shortValues},
			})
			ts := httptest.NewServer(r)
			defer ts.Close()

			requestPath, err := url.JoinPath(ts.URL, "api", "shorten", "batch")
			if err != nil {
				errCh <- err
				return
			}

			req, err := http.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
			if err != nil {
				errCh <- err
				return
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusCreated {
				errCh <- fmt.Errorf("unexpected status %d", res.StatusCode)
				return
			}

			var batchResp model.ShortenBatchResp
			if err := json.NewDecoder(res.Body).Decode(&batchResp); err != nil {
				errCh <- err
				return
			}

			if len(batchResp) != urlsPerWorker {
				errCh <- fmt.Errorf("expected %d urls, got %d", urlsPerWorker, len(batchResp))
				return
			}

			deleteKeys := make([]string, 0, urlsPerWorker)
			shortRefs := make([]struct {
				resolveURL string
				shortKey   string
			}, 0, urlsPerWorker)

			for _, item := range batchResp {
				parsed, err := url.Parse(item.ShortURL)
				if err != nil {
					errCh <- err
					return
				}

				shortKey := strings.TrimPrefix(parsed.Path, "/")
				deleteKeys = append(deleteKeys, shortKey)
				shortRefs = append(shortRefs, struct {
					resolveURL string
					shortKey   string
				}{resolveURL: ts.URL + parsed.Path, shortKey: shortKey})
			}

			deleteBody, err := json.Marshal(deleteKeys)
			if err != nil {
				errCh <- err
				return
			}

			deletePath, err := url.JoinPath(ts.URL, "api", "user", "urls")
			if err != nil {
				errCh <- err
				return
			}

			deleteReq, err := http.NewRequest(http.MethodDelete, deletePath, bytes.NewReader(deleteBody))
			if err != nil {
				errCh <- err
				return
			}
			deleteReq.Header.Set("Content-Type", "application/json")

			deleteRes, err := client.Do(deleteReq)
			if err != nil {
				errCh <- err
				return
			}
			deleteRes.Body.Close()

			if deleteRes.StatusCode != http.StatusAccepted {
				errCh <- fmt.Errorf("unexpected delete status %d", deleteRes.StatusCode)
				return
			}

			for _, ref := range shortRefs {
				if err := waitForGone(client, ref.resolveURL, ref.shortKey); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}
