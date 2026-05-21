package handler_test

import (
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

func (s *HandlerSuite) getRouter(testStructure *innerStructure) *chi.Mux {
	if testStructure == nil {
		testStructure = &innerStructure{}
	}

	if testStructure.store == nil {
		testStructure.store = repository.NewInMemoryStore(nil)
	}

	ctrl := gomock.NewController(s.T())
	if testStructure.strGen == nil {
		testStructure.strGen = newMockStringGenerator(ctrl)
	}

	log := zerolog.Nop()
	if testStructure.config == nil {
		cfg, err := config.GetConfig([]string{}, &log)
		s.Require().NoError(err)
		testStructure.config = cfg
	}

	if testStructure.config.RedirectDomain.Host == "" {
		testStructure.config.RedirectDomain = testHost
	}

	mockAudit := auditmocks.NewMockPublisher(ctrl)
	mockAudit.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sf, err := shortifier.NewShortifier(s.T().Context(), testStructure.strGen, testStructure.store, mockAudit, &testStructure.config.RedirectDomain)
	s.Require().NoError(err)

	h := handler.NewHandlers(sf, testStructure.store, mockAudit)

	return router.NewRouter(h)
}
