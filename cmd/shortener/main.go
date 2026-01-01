package main

import (
	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.GetConfig()

	store := repository.NewStore()
	hashFunc := hasher.NewHasher("SomeSecret")
	fizzBuzzService := shortifier.NewShortifier(hashFunc, store)

	return handler.Serve(cfg.Handlers, fizzBuzzService)
}
