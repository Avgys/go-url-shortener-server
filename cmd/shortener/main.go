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
	shutdownServerLimit = 30 * time.Second
	shutdownLimit       = 10 * time.Second
)

var (
	BuildVersion    string = "N/A"
	BuildDate       string = "N/A"
	BuildCommitHash string = "N/A"
	BuildCommitName string = "N/A"
	BuildBranch     string = "N/A"
)

func main() {
	log, closeLogger, err := logger.NewBaseLogger(logger.GetFuncName())
	if err != nil {
		fmt.Println("failed to create logger", err)
		panic(err)
	}

	fillBuildInfoFromGit()
	printBuildInfo()

	defer func() { _ = closeLogger() }()

	if err := run(log); err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Println("bye-bye")
}

func run(log *zerolog.Logger) error {

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGQUIT)
	defer stop()

	g, ctx := errgroup.WithContext(rootCtx)

	srv, err := server.NewServer(ctx, log)

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

		if err := srv.Start(); err != nil {
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

		shutdownErr := srv.Shutdown(shutdownTimeoutCtx)
		if shutdownErr != nil {
			log.Printf("an error occurred during server shutdown: %v", shutdownErr)
		}

		return shutdownErr
	})

	if err := g.Wait(); err != nil {
		log.Err(err).Send()
		return err
	}

	return nil
}
