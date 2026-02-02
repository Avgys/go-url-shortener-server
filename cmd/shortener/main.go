package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(); err != nil {
		logger.Log.Fatal().
			Err(err)
	}

	<-ctx.Done()

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
