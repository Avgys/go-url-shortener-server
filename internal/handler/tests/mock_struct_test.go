package handler_test

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go-url-shortener/internal/config"
	flagvalues "go-url-shortener/internal/config/flag_values"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	"go-url-shortener/internal/service"
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
