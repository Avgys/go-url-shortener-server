package handler_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"go-url-shortener/internal/config"
	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	auditmocks "go-url-shortener/internal/service/audit/mocks"
	"go-url-shortener/internal/service/shortifier"
	shortifiermocks "go-url-shortener/internal/service/shortifier/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
)

var testHost = flagvalues.NetAddress{Host: "localhost:8080", Scheme: "http", SchemeRequired: true}

type innerStructure struct {
	store  repository.Repository
	strGen shortifier.StringGenerator
	config *config.Config
}

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

func newMockStringGenerator(ctrl *gomock.Controller) shortifier.StringGenerator {
	m := shortifiermocks.NewMockStringGenerator(ctrl)
	m.EXPECT().GetRandomString(gomock.Any()).Return("test-str").AnyTimes()
	return m
}

// benchStrGen returns unique short hashes so benchmarks do not exhaust collision retries.
type benchStrGen struct {
	seq atomic.Uint64
}

func (g *benchStrGen) GetRandomString(int) string {
	return fmt.Sprintf("bench%x", g.seq.Add(1))
}

func (s *HandlerSuite) getRouter(testStructure *innerStructure) *chi.Mux {
	return buildRouter(s.T(), testStructure)
}

func buildRouter(tb testing.TB, testStructure *innerStructure) *chi.Mux {
	tb.Helper()

	if testStructure == nil {
		testStructure = &innerStructure{}
	}

	if testStructure.store == nil {
		testStructure.store = repository.NewInMemoryStore(nil)
	}

	ctrl := gomock.NewController(tb)
	tb.Cleanup(func() { ctrl.Finish() })

	if testStructure.strGen == nil {
		testStructure.strGen = newMockStringGenerator(ctrl)
	}

	log := zerolog.Nop()
	if testStructure.config == nil {
		cfg, err := config.GetConfig([]string{}, &log)
		if err != nil {
			tb.Fatal(err)
		}
		testStructure.config = cfg
	}

	if testStructure.config.RedirectDomain.Host == "" {
		testStructure.config.RedirectDomain = testHost
	}

	mockAudit := auditmocks.NewMockPublisher(ctrl)
	mockAudit.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	ctx := context.Background()
	if c, ok := tb.(interface{ Context() context.Context }); ok {
		ctx = c.Context()
	}

	sf, err := shortifier.NewShortifier(ctx, testStructure.strGen, testStructure.store, mockAudit, &testStructure.config.RedirectDomain)
	if err != nil {
		tb.Fatal(err)
	}

	h := handler.NewHandlers(sf, testStructure.store, mockAudit, nil)

	return router.NewRouter(h)
}
