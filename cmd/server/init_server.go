package server

import (
	"context"
	"net/http"
	"os"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/rs/zerolog"
)

func GetServer(done context.Context, traceLogger *zerolog.Logger) (*http.Server, error) {

	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	h, err := prepareDI(done, cfg, traceLogger)

	r := router.NewRouter(h)

	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv, nil
}

func prepareDI(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) (*handler.Handlers, error) {

	store, err := repository.NewRepository(done, cfg, traceLogger)

	if err != nil {
		traceLogger.Err(err).Msg("error initializing repository")
		return nil, err
	}

	generator := service.NewStringGenerator()
	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)
	h := handler.NewHandlers(shortifier, store)

	return h, nil
}
