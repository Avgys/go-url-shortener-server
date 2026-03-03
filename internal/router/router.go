package router

import (
	auth_middlewares "github.com/Avgys/go-url-shortener-server/internal/auth/middlewares"
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

	r.Group(func(r chi.Router) {

		r.Use(auth_middlewares.SetCookie)
		r.With(middleware.AllowContentType(textType, xgzipType)).Post("/", h.ShortifyURL)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AllowContentType(jsonType))

			r.Post("/api/shorten", h.ShortenURL)
			r.Post("/api/shorten/batch", h.ShortenBatch)
		})
	})

	r.With(auth_middlewares.SetCookie, auth_middlewares.RequireCookie).Get("/api/user/urls", h.GetURLsByUserID)
	r.Get("/{url}", h.Redirect)
	r.Get("/ping", h.Ping)

	return r
}
