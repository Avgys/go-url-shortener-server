package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceLogger := logger.Endpoint(ctx, "Redirect")

	url := chi.URLParam(r, "url")
	url, err := h.Shortifier.ResolveShortURL(ctx, url, traceLogger)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Location", url)
	shared.WriteResponse(w, nil, http.StatusTemporaryRedirect)
}

func (h *Handlers) GetURLsByUserId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "GetURLsByUserId")

	claims, exist := getClaims(ctx)
	if !exist {
		shared.WriteResponse(w, nil, http.StatusUnauthorized)
		return
	}

	urls, err := h.Shortifier.GetURLsByUserId(ctx, claims.UserID, traceLogger)
	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	response, _ := json.Marshal(urls)

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, response, http.StatusOK)
}
