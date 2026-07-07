package main

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/server"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

func runHTTPServer(rootCtx context.Context, log *zerolog.Logger, cfg *config.Config, handlers *handler.Handlers) error {
	g, ctx := errgroup.WithContext(rootCtx)

	srv, err := server.NewServer(ctx, cfg, handlers, log)

	if err != nil {
		return err
	}

	shutdownDone := make(chan struct{})

	// Enforce app shutdown
	go func() {
		<-ctx.Done()
		timer := time.NewTimer(shutdownLimit)
		defer timer.Stop()

		select {
		case <-shutdownDone:
			return
		case <-timer.C:
			log.Fatal().Msg("failed to gracefully shutdown the service")
		}
	}()

	g.Go(func() error { return startServer(srv) })
	g.Go(func() error { return gracefulShutdownServer(ctx, log, shutdownDone, srv) })

	if err := g.Wait(); err != nil {
		log.Err(err).Send()
		return err
	}
	return nil
}

func startServer(srv *server.Server) (err error) {
	defer func() {
		errRec := recover()
		if errRec != nil {
			err = fmt.Errorf("a panic occurred: %v", errRec)
		}
	}()

	if err := srv.Start(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen and server has failed: %w", err)
	}

	return err
}

func gracefulShutdownServer(ctx context.Context, log *zerolog.Logger, shutdownDone chan struct{}, srv *server.Server) error {
	defer log.Print("server has been shutdown")

	<-ctx.Done()
	defer close(shutdownDone)

	shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), shutdownServerLimit)
	defer cancelShutdownTimeoutCtx()

	shutdownErr := srv.Shutdown(shutdownTimeoutCtx)
	if shutdownErr != nil {
		log.Printf("an error occurred during server shutdown: %v", shutdownErr)
	}

	return shutdownErr
}
