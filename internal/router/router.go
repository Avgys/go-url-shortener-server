package router

import (
	"github.com/Avgys/go-url-shortener-server/internal/handler"
	"github.com/Avgys/go-url-shortener-server/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *handler.Handlers) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.RealIP, middlewares.WithLogging)

	r.With(middleware.AllowContentType("text/plain")).Post("/", h.ShortifyURL)
	r.Get("/{url}", h.Redirect)

	return r
}
