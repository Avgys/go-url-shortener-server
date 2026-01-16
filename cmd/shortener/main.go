package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.GetConfig(os.Args[1:])

	if err != nil {
		return err
	}

	r := prepareRouter(cfg)

	srv := &http.Server{
		Addr:    cfg.AppURL.Host,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func prepareRouter(cfg *config.Config) *chi.Mux {
	store := repository.NewStore(nil)
	generator := service.NewStringGenerator()

	shortifier := service.NewShortifier(generator, store, &cfg.RedirectDomain)

	h := handler.NewHandlers(shortifier)
	return router.NewRouter(h)
}
