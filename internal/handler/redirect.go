package handler

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	ctx := r.Context()

	traceLogger := logger.Endpoint(ctx, "Redirect")

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(ctx, url, traceLogger); err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Location", url)
	shared.WriteResponse(w, nil, http.StatusTemporaryRedirect)
}
