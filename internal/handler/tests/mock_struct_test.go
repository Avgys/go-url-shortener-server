package handler_test

import (
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	flagvalues "github.com/Avgys/go-url-shortener-server/internal/config/flag_values"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

var testHost flagvalues.NetAddress = flagvalues.NetAddress{Host: "localhost:8080", Scheme: "http"}

type innerStructure struct {
	store      repository.Repository
	strGen     service.StringGenerator
	shortifier *service.Shortifier
	handlers   *handler.Handlers
	config     *config.Config
}

func getRouter(t *testing.T, testStructure *innerStructure) *chi.Mux {
	t.Helper()

	if testStructure == nil {
		testStructure = &innerStructure{}
	}

	if testStructure != nil && testStructure.store == nil {
		testStructure.store = repository.NewInMemoryStore(nil)
	}

	if testStructure != nil && testStructure.strGen == nil {
		testStructure.strGen = &service.StrGenerator{}
	}

	if testStructure != nil && testStructure.config == nil {
		testStructure.config, _ = config.GetConfig([]string{}, &zerolog.Logger{})
	}

	shortifier := service.NewShortifier(t.Context(), testStructure.strGen, testStructure.store, &testStructure.config.RedirectDomain)

	h := &handler.Handlers{
		Shortifier: shortifier,
	}

	return router.NewRouter(h)
}
