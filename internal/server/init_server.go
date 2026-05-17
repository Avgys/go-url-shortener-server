package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	"go-url-shortener/internal/service"
	"go-url-shortener/internal/service/shortifier"

	"github.com/rs/zerolog"
)

func NewServer(done context.Context, traceLogger *zerolog.Logger) (*http.Server, error) {

	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	h, err := prepareDI(done, cfg, traceLogger)

	if err != nil {
		return nil, err
	}

	r := router.NewRouter(h)

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv, nil
}

func prepareDI(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) (*handler.Handlers, error) {

	closers := make([]io.Closer, 0)

	go func() {
		<-done.Done()

		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				traceLogger.Err(err).Str("message", "error when closing services")
			}
		}
	}()

	store, err := repository.NewRepository(done, cfg, traceLogger)

	if err != nil {
		return nil, fmt.Errorf("error initializing repository: %w", err)
	}

	if closer, ok := store.(io.Closer); ok {
		closers = append(closers, closer)
	}

	generator := service.NewStringGenerator()
	shortifierService, err := shortifier.NewShortifier(done, generator, store, &cfg.RedirectDomain)
	if err != nil {
		return nil, fmt.Errorf("error initializing shortifier: %w", err)
	}
	h := handler.NewHandlers(shortifierService, store)

	return h, nil
}
