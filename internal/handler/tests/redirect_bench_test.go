package handler_test

import (
	"fmt"
	dbmodel "go-url-shortener/internal/model/db"
	"go-url-shortener/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"
)

const benchRedirectEntries = 1024

func benchRedirectStore() repository.Repository {
	data := make([]*dbmodel.DBURL, benchRedirectEntries)
	for i := range data {
		data[i] = &dbmodel.DBURL{
			OriginalURL: benchmarkLongURL(i),
			ShortURL:    fmt.Sprintf("r%x", i),
		}
	}
	return repository.NewInMemoryStore(data)
}

func BenchmarkRedirect(b *testing.B) {
	setupBench(b)
	const host = "http://localhost:8080"

	r := buildRouter(b, &innerStructure{store: benchRedirectStore()})

	b.ResetTimer()
	b.ReportAllocs()

	i := 0
	for b.Loop() {
		b.StopTimer()
		i++
		requestURL := fmt.Sprintf("%s/r%x", host, i%benchRedirectEntries)
		req := httptest.NewRequest(http.MethodGet, requestURL, nil)

		rec := httptest.NewRecorder()

		b.StartTimer()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusTemporaryRedirect {
			b.Fatalf("unexpected status %d", rec.Code)
		}
		wantLocation := benchmarkLongURL(i % benchRedirectEntries)
		if got := rec.Header().Get("Location"); got != wantLocation {
			b.Fatalf("unexpected location %q, want %q", got, wantLocation)
		}
	}
}
