package main

import (
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
)

var closers = make([]io.Closer, 0)

func main() {

	if err := run(); err != nil {
		logger.DefaulLogger.Fatal().
			Err(err)
	}

	for _, closer := range closers {
		closer.Close()
	}
}

func run() error {
	cfg, err := config.GetConfig(os.Args[1:])

	if err != nil {
		return err
	}

	r, err := prepareRouter(cfg)

	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func prepareRouter(cfg *config.Config) (*chi.Mux, error) {
	store, err := repository.NewFileStore(string(cfg.FileStorage))

	if err != nil {
		return nil, err
	}

	generator := service.NewStringGenerator()

	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)

	h := handler.NewHandlers(shortifier)
	return router.NewRouter(h), nil
}
