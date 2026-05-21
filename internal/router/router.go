package router

import (
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/middlewares"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const textType = "text/plain"
const xgzipType = "application/x-gzip"
const jsonType = "application/json"

func NewRouter(h *handler.Handlers) *chi.Mux {

	r := chi.NewRouter()
	setEndpoints(r, h)

	return r
}

func setEndpoints(r *chi.Mux, h *handler.Handlers) {

	r.Use(middleware.RealIP, middlewares.Recoverer, middlewares.WithLogging, middlewares.WithCompression)

	r.Group(func(r chi.Router) {

		r.Use(middlewares.SetCookie)
		r.With(middleware.AllowContentType(textType, xgzipType)).Post("/", h.ShortifyURL)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AllowContentType(jsonType))

			r.Post("/api/shorten", h.ShortenURL)
			r.Post("/api/shorten/batch", h.ShortenBatch)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.SetCookie, middlewares.AuthRequireCookie(true))

		r.Route("/api/user", func(r chi.Router) {
			r.Get("/urls", h.GetURLsByUserID)
			r.With(middleware.AllowContentType(jsonType)).Delete("/urls", h.DeleteShortURL)
		})
	})

	r.With(middlewares.AuthRequireCookie(false)).Get("/{url}", h.Redirect)
	r.Get("/ping", h.Ping)
}
