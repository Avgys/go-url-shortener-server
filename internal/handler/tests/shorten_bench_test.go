package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/model/requests"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func init() {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.bench") {
			logger.SetDiscardOutput(true)
			return
		}
	}
}

func benchmarkLongURL(i int) string {
	return fmt.Sprintf("http://long-url-%d.com", i)
}

func BenchmarkShortifyURL(b *testing.B) {
	const host = "http://localhost:8080"

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, host, strings.NewReader(benchmarkLongURL(i)))
		req.Header.Set("Content-Type", "text/plain")

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}

func BenchmarkShortenURL(b *testing.B) {
	const host = "http://localhost:8080"

	requestPath, err := url.JoinPath(host, "api", "shorten")
	if err != nil {
		b.Fatal(err)
	}

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonBody, err := json.Marshal(requests.ShortenReq{URL: benchmarkLongURL(i)})
		if err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}

func BenchmarkShortenBatch(b *testing.B) {
	const host = "http://localhost:8080"

	requestPath, err := url.JoinPath(host, "api", "shorten", "batch")
	if err != nil {
		b.Fatal(err)
	}

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		payload, err := json.Marshal(requests.ShortenBatchReq{
			{CorrelationID: "1", FullURL: benchmarkLongURL(i * 2)},
			{CorrelationID: "2", FullURL: benchmarkLongURL(i*2 + 1)},
		})
		if err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}
