package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
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
			Err(err).
			Send()
	}
}

func run(traceLogger *zerolog.Logger) error {
	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return err
	}

	r, err, closers := prepareRouter(cfg, traceLogger)

	defer func() {
		for _, closer := range closers {
			closer.Close()
		}
	}()

	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func prepareRouter(cfg *config.Config, traceLogger *zerolog.Logger) (*chi.Mux, error, []io.Closer) {

	closers := make([]io.Closer, 0)

	store, err := repository.NewRepository(context.Background(), cfg)

	if err != nil {
		traceLogger.Err(err).Msg("error initializing repository")
		return nil, err, closers
	}

	closers = append(closers, store)

	generator := service.NewStringGenerator()

	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)

	h := handler.NewHandlers(shortifier, store)
	return router.NewRouter(h), nil, closers
}
