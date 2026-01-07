package main

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/router"
	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	r := prepareRouter()

	srv := &http.Server{
		Addr:    cfg.URL.Host,
		Handler: r,
	}

	return srv.ListenAndServe()
}

func prepareRouter() *chi.Mux {
	store := repository.NewStore()
	hashFunc := hasher.NewHasher("SomeSecret")

	shortifier := shortifier.NewShortifier(hashFunc, store)

	h := handler.NewHandlers(shortifier)
	return router.NewRouter(h)
}
