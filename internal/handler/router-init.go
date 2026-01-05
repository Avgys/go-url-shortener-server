package handler

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/handler/config"
)

type Shortifier interface {
	ResolveShortURL(shortURL string) (string, error)
	ShortifyURL(url string) (string, bool, error)
}

func Serve(cfg config.Config, shortifier Shortifier) error {

	h := &Handlers{Shortifier: shortifier}

	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.ShortifyURL)
	mux.HandleFunc("GET /{url}", h.Redirect)

	return mux
}
