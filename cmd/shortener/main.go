package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/repository/db"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {

	logger := logger.NewLogger().
		With().
		Str("component", "initialize").
		Logger()

	if err := run(&logger); err != nil {
		logger.Fatal().
			Err(err)
	}
}

func run(traceLogger *zerolog.Logger) error {
	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return err
	}

	r, err := prepareRouter(cfg, traceLogger)

	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func prepareRouter(cfg *config.Config, traceLogger *zerolog.Logger) (*chi.Mux, error) {
	store, err := repository.NewDBStore(context.Background(), &db.Config{ConnectionString: cfg.DBConnectionString})

	if err != nil {
		return nil, err
	}

	generator := service.NewStringGenerator()

	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)

	h := handler.NewHandlers(shortifier, store)
	return router.NewRouter(h), nil
}
