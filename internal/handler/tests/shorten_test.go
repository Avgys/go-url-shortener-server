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
	"testing"
	"time"

	dbmodel "go-url-shortener/internal/model/db"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/model/responses"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/testcommon"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

type HandlerSuite struct {
	suite.Suite
}

func (s *HandlerSuite) Test_handlers_ShortifyURL() {
	host := "http://localhost:8080"
	awaitedStr, err := url.JoinPath(host, "test-str")
	s.Require().NoError(err)

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name:             "create shorturl",
			url:              "http://long-url.com",
			defaultStructure: &innerStructure{},
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
				store: repository.NewInMemoryStore([]*dbmodel.DBURL{{OriginalURL: "http://long-url.com", ShortURL: "short-hash"}}),
			},
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
		s.Run(tt.name, func() {

			//Init
			req := httptest.NewRequest(http.MethodPost, host, strings.NewReader(tt.url))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "text/plain")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := s.getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer func() { _ = res.Body.Close() }()

			//Check

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}

func (s *HandlerSuite) Test_handlers_CreateShortURLAndRead() {
	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name:             "create shorturl and read",
			url:              "http://long-url.com",
			defaultStructure: &innerStructure{},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusTemporaryRedirect,
				Headers:    map[string]string{"Location": "http://long-url.com"},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {

			//Init
			r := s.getRouter(tt.defaultStructure)

			//Get short url
			req := httptest.NewRequest(http.MethodPost, testHost.String(), strings.NewReader(tt.url))
			req.Header.Set("Content-Type", "text/plain")
			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			resBody, err := io.ReadAll(res.Body)
			_ = res.Body.Close()

			s.Require().NoError(err)
			s.Require().Equal(http.StatusCreated, res.StatusCode)

			shortURL := string(resBody)

			//Resolve short url
			req = httptest.NewRequest(http.MethodGet, shortURL, nil)
			req.SetPathValue("url", "short-hash")
			recorder = httptest.NewRecorder()
			r.ServeHTTP(recorder, req)
			res = recorder.Result()
			defer func() { _ = res.Body.Close() }()

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}

func (s *HandlerSuite) Test_handlers_ShortenURL() {
	host := "http://localhost:8080"
	requestPath, err := url.JoinPath(host, "api", "shorten")
	s.Require().NoError(err)

	awaitedStr, err := url.JoinPath(host, "test-str")
	s.Require().NoError(err)

	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name:             "create shorturl",
			url:              "http://long-url.com",
			defaultStructure: &innerStructure{},
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
				store: repository.NewInMemoryStore([]*dbmodel.DBURL{{OriginalURL: "http://long-url.com", ShortURL: "short-hash"}}),
			},
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
		s.Run(tt.name, func() {

			//Init

			jsonBody, err := json.Marshal(requests.ShortenReq{URL: tt.url})
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := s.getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer func() { _ = res.Body.Close() }()

			//Check

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}

func (s *HandlerSuite) Test_handlers_ShortenBatch() {
	host := "http://localhost:8080"
	requestPath, err := url.JoinPath(host, "api", "shorten", "batch")
	s.Require().NoError(err)

	shortURL1, err := url.JoinPath(host, "short-1")
	s.Require().NoError(err)

	shortURL2, err := url.JoinPath(host, "short-2")
	s.Require().NoError(err)

	tests := []struct {
		name             string
		payload          requests.ShortenBatchReq
		defaultStructure *innerStructure
		contentType      string
		want             testcommon.ResponseWant
	}{
		{
			name: "create short urls",
			payload: requests.ShortenBatchReq{
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
			payload: requests.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "http://long-url.com"},
			},
			defaultStructure: &innerStructure{
				strGen: &mockStrGenSequence{values: []string{"short-1"}},
				store:  repository.NewInMemoryStore([]*dbmodel.DBURL{&dbmodel.DBURL{OriginalURL: "http://long-url.com", ShortURL: "short-1"}}),
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
			payload: requests.ShortenBatchReq{
				{CorrelationID: "1", FullURL: "/gdfgdfhs"},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty url",
			payload: requests.ShortenBatchReq{
				{CorrelationID: "1", FullURL: ""},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "wrong content-type",
			payload: requests.ShortenBatchReq{
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
		s.Run(tt.name, func() {

			jsonBody, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))

			if tt.contentType == "" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}

			recorder := httptest.NewRecorder()
			r := s.getRouter(tt.defaultStructure)

			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer func() { _ = res.Body.Close() }()

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}

func (s *HandlerSuite) Test_handlers_ShortenBatchResolveByCorrelation() {
	host := "http://localhost:8080"
	requestPath, _ := url.JoinPath(host, "api", "shorten", "batch")

	source := map[string]string{
		"1": "http://long-url-1.com",
		"2": "http://long-url-2.com",
	}

	payload := requests.ShortenBatchReq{
		{CorrelationID: "1", FullURL: source["1"]},
		{CorrelationID: "2", FullURL: source["2"]},
	}

	jsonBody, err := json.Marshal(payload)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	r := s.getRouter(&innerStructure{
		strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
	})

	r.ServeHTTP(recorder, req)
	res := recorder.Result()
	defer func() { _ = res.Body.Close() }()

	s.Require().Equal(http.StatusCreated, res.StatusCode)

	var batchResp responses.ShortenBatchResp
	err = json.NewDecoder(res.Body).Decode(&batchResp)
	s.Require().NoError(err)

	for _, item := range batchResp {
		original, ok := source[item.CorrelationID]
		s.Require().True(ok)

		parsed, err := url.Parse(item.ShortURL)
		s.Require().NoError(err)
		shortKey := strings.TrimPrefix(parsed.Path, "/")

		resolveReq := httptest.NewRequest(http.MethodGet, item.ShortURL, nil)
		resolveReq.SetPathValue("url", shortKey)
		resolveRecorder := httptest.NewRecorder()
		r.ServeHTTP(resolveRecorder, resolveReq)
		resolveRes := resolveRecorder.Result()
		_ = resolveRes.Body.Close()

		s.Require().Equal(http.StatusTemporaryRedirect, resolveRes.StatusCode)
		s.Require().Equal(original, resolveRes.Header.Get("Location"))
	}
}

func (s *HandlerSuite) Test_handlers_ShortenDeleteRead() {
	const workerCount = 1
	const urlsPerWorker = 3

	r := s.getRouter(&innerStructure{
		strGen: &mockStrGenSequence{
			values: []string{"del-1", "del-2", "del-3", "del-4", "del-5", "del-6", "del-7", "del-8", "del-9", "del-10"},
		},
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

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
			_ = resolveRes.Body.Close()

			if resolveRes.StatusCode == http.StatusGone {
				return nil
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("expected %d, got %d", http.StatusGone, resolveRes.StatusCode)
			}

			time.Sleep(20 * time.Millisecond)
		}
	}

	var g errgroup.Group

	for i := 0; i < workerCount; i++ {
		idx := i
		g.Go(func() error {

			jar, err := cookiejar.New(nil)
			if err != nil {
				return err
			}

			client := &http.Client{
				Jar: jar,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			payload := requests.ShortenBatchReq{
				{CorrelationID: "1", FullURL: fmt.Sprintf("http://long-url-%d-1.com", idx)},
				{CorrelationID: "2", FullURL: fmt.Sprintf("http://long-url-%d-2.com", idx)},
				{CorrelationID: "3", FullURL: fmt.Sprintf("http://long-url-%d-3.com", idx)},
			}

			jsonBody, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			requestPath, err := url.JoinPath(ts.URL, "api", "shorten", "batch")
			if err != nil {
				return err
			}

			req, err := http.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = res.Body.Close() }()

			if res.StatusCode != http.StatusCreated {
				return fmt.Errorf("unexpected status %d", res.StatusCode)
			}

			var batchResp responses.ShortenBatchResp
			if err := json.NewDecoder(res.Body).Decode(&batchResp); err != nil {
				return err
			}

			if len(batchResp) != urlsPerWorker {
				return fmt.Errorf("expected %d urls, got %d", urlsPerWorker, len(batchResp))
			}

			deleteKeys := make([]string, 0, urlsPerWorker)
			shortRefs := make([]struct {
				resolveURL string
				shortKey   string
			}, 0, urlsPerWorker)

			for _, item := range batchResp {
				parsed, err := url.Parse(item.ShortURL)
				if err != nil {
					return err
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
				return err
			}

			deletePath, err := url.JoinPath(ts.URL, "api", "user", "urls")
			if err != nil {
				return err
			}

			deleteReq, err := http.NewRequest(http.MethodDelete, deletePath, bytes.NewReader(deleteBody))
			if err != nil {
				return err
			}
			deleteReq.Header.Set("Content-Type", "application/json")

			deleteRes, err := client.Do(deleteReq)
			if err != nil {
				return err
			}
			_ = deleteRes.Body.Close()

			if deleteRes.StatusCode != http.StatusAccepted {
				return fmt.Errorf("unexpected delete status %d", deleteRes.StatusCode)
			}

			for _, ref := range shortRefs {
				if err := waitForGone(client, ref.resolveURL, ref.shortKey); err != nil {
					return err
				}
			}

			return nil
		})
	}

	s.Require().NoError(g.Wait())
}
