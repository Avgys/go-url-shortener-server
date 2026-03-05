package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	stdlog "log"

	"github.com/Avgys/go-url-shortener-server/cmd/server"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"golang.org/x/sync/errgroup"
)

const (
	shutdownServerLimit = 5
	shutdownLimit       = 10
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

	log.Println("bye-bye")
}

func run() error {

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL, os.Interrupt)
	defer stop()

	log, close := logger.NewLogger()
	defer close()

	g, ctx := errgroup.WithContext(rootCtx)

	*log = log.With().
		Str("component", "initialize").
		Logger()

	srv, err := server.GetServer(ctx, log)

	if err != nil {

		log.Error().
			Err(err).
			Send()

		os.Exit(1)
	}

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), shutdownLimit)
		defer cancelCtx()

		<-ctx.Done()
		stdlog.Fatal("failed to gracefully shutdown the service")
	})

	g.Go(func() (err error) {
		defer func() {
			errRec := recover()
			if errRec != nil {
				err = fmt.Errorf("a panic occurred: %v", errRec)
			}
		}()

		if err = srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			return fmt.Errorf("listen and server has failed: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		defer log.Print("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), shutdownServerLimit)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Err(err).Send()
	}

	return err
}
