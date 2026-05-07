package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/db"
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

	//Db
	dbConnection, err := db.NewDB(done, &db.Config{ConnectionString: cfg.DBConnectionString})

	closers = append(closers, dbConnection)

	go func() {
		<-done.Done()

		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				traceLogger.Err(err).Str("message", "error when closing services")
			}
		}
	}()

	if err != nil {
		err = fmt.Errorf("error initializing db: %w", err)
		return nil, err
	}

	store, err := repository.NewRepository(done, cfg, traceLogger)

	if err != nil {
		err = fmt.Errorf("error initializing repository: %w", err)
		return nil, err
	}

	generator := service.NewStringGenerator()
	shortifier := service.NewShortifier(done, generator, store, &cfg.RedirectDomain)
	h := handler.NewHandlers(shortifier, store)

	return h, nil
}
