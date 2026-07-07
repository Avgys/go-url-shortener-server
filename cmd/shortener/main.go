package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
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

	cfg, err := config.GetConfig(os.Args[1:], log)
	if err != nil {
		return err
	}

	handlers, err := server.PrepareHandlers(rootCtx, cfg, log)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(rootCtx)
	g.Go(func() error { return runHTTPServer(ctx, log, cfg, handlers) })
	g.Go(func() error { return runGRPCServer(ctx, log, cfg, handlers) })

	return g.Wait()
}

func runGRPCServer(rootCtx context.Context, log *zerolog.Logger, cfg *config.Config, handlers *handler.Handlers) error {
	listenAddr := GRPCListenAddr(cfg.GRPCPort)
	log.Print("grpc server started on " + listenAddr)
	return ServeGRPC(rootCtx, handlers.Shortifier, listenAddr)
}
