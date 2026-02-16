package main

import (
	"context"
	"errors"
	"io"
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
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	log, close := logger.NewLogger()

	*log = log.With().
		Str("component", "initialize").
		Logger()

	defer close()

	srv, err := run(log)

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
		// close db, etc.
		os.Exit(0)
		return // allow normal exit code 0
	case err := <-errCh:
		if err != nil {
			// startup/runtime error -> exit non-zero
			log.Println("server error:", err)
			os.Exit(1)
		}
	}
}

func run(traceLogger *zerolog.Logger) (*http.Server, error) {
	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	r, err, _ := prepareRouter(cfg, traceLogger)

	// defer func() {
	// 	for _, closer := range closers {
	// 		closer.Close()
	// 	}
	// }()

	if err != nil {
		return nil, err
	}

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv, nil
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
