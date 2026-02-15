package router

import (
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const textType = "text/plain"
const xgzipType = "application/x-gzip"
const jsonType = "application/json"

func NewRouter(h *handler.Handlers) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.RealIP, middlewares.WithLogging, middlewares.WithCompression)

	r.With(middleware.AllowContentType(textType, xgzipType)).Post("/", h.ShortifyURL)
	r.With(middleware.AllowContentType(jsonType)).Post("/api/shorten", h.ShortenURL)
	r.Get("/{url}", h.Redirect)
	r.Get("/ping", h.Ping)

	return r
}
