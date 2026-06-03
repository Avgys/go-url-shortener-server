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
	"strings"
	"testing"
)

func setupBench(b *testing.B) {
	b.Helper()
	logger.SetDiscardOutput(true)
	b.Cleanup(func() { logger.SetDiscardOutput(false) })
}

func benchmarkLongURL(i int) string {
	return fmt.Sprintf("http://long-url-%d.com", i)
}

func BenchmarkShortifyURL(b *testing.B) {
	setupBench(b)
	const host = "http://localhost:8080"

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ReportAllocs()

	i := 0
	for b.Loop() {
		b.StopTimer()
		i++
		req := httptest.NewRequest(http.MethodPost, host, strings.NewReader(benchmarkLongURL(i)))
		req.Header.Set("Content-Type", "text/plain")

		rec := httptest.NewRecorder()
		b.StartTimer()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}

func BenchmarkShortenURL(b *testing.B) {
	setupBench(b)
	const host = "http://localhost:8080"

	requestPath, err := url.JoinPath(host, "api", "shorten")
	if err != nil {
		b.Fatal(err)
	}

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		b.StopTimer()
		i++

		jsonBody, err := json.Marshal(requests.ShortenReq{URL: benchmarkLongURL(i)})
		if err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, requestPath, bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		b.StartTimer()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}

func BenchmarkShortenBatch(b *testing.B) {
	setupBench(b)
	const host = "http://localhost:8080"

	requestPath, err := url.JoinPath(host, "api", "shorten", "batch")
	if err != nil {
		b.Fatal(err)
	}

	r := buildRouter(b, &innerStructure{strGen: &benchStrGen{}})

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		b.StopTimer()
		i++
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
		b.StartTimer()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}
