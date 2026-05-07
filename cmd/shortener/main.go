package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/server"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	shutdownServerLimit = 5 * time.Second
	shutdownLimit       = 10 * time.Second
)

func main() {

	log, closeLogger := logger.NewLogger()
	defer closeLogger()

	if err := run(log); err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Println("bye-bye")
}

func run(log *zerolog.Logger) error {

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()

	*log = log.With().
		Str("component", "initialize").
		Logger()

	g, ctx := errgroup.WithContext(rootCtx)

	srv, err := server.GetServer(ctx, log)

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

	// start server
	g.Go(func() (err error) {
		defer func() {
			errRec := recover()
			if errRec != nil {
				err = fmt.Errorf("a panic occurred: %v", errRec)
			}
		}()

		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return fmt.Errorf("listen and server has failed: %w", err)
		}

		return err
	})

	// graceful shutdown
	g.Go(func() error {
		defer log.Print("server has been shutdown")

		<-ctx.Done()
		defer close(shutdownDone)

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), shutdownServerLimit)
		defer cancelShutdownTimeoutCtx()

		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}

		return err
	})

	if err := g.Wait(); err != nil {
		log.Err(err).Send()
		return err
	}

	return nil
}
