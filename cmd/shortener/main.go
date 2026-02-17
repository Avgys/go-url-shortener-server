package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/Avgys/go-url-shortener-server/internal/service/closer"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	aggCloser := closer.NewCloser()

	log, close := logger.NewLogger()
	defer close()

	*log = log.With().
		Str("component", "initialize").
		Logger()

	srv, err := getServer(log, aggCloser)

	if err != nil {

		log.Error().
			Err(err).
			Send()

		os.Exit(1)
	}

	errCh := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// graceful shutdown
		shutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_ = srv.Shutdown(shutCtx)

		if err := aggCloser.Close(); err != nil {
			log.Err(err).Send()
		}

		// close db, etc.
		return // allow normal exit code 0
	case err := <-errCh:
		if err != nil {
			// startup/runtime error -> exit non-zero

			log.Err(err).Msg("server error:")

			if err := aggCloser.Close(); err != nil {
				log.Err(err).Send()
			}

			os.Exit(1)
		}
	}
}

func getServer(traceLogger *zerolog.Logger, aggCloser *closer.Closer) (*http.Server, error) {

	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	r, err := prepareRouter(cfg, aggCloser, traceLogger)

	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv, nil
}

func prepareRouter(cfg *config.Config, aggCloser *closer.Closer, traceLogger *zerolog.Logger) (*chi.Mux, error) {

	store, err := repository.NewRepository(context.Background(), cfg)
	aggCloser.Add(store)

	if err != nil {
		traceLogger.Err(err).Msg("error initializing repository")
		return nil, err
	}

	generator := service.NewStringGenerator()
	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)
	h := handler.NewHandlers(shortifier, store)

	return router.NewRouter(h), nil
}
