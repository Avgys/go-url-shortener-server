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
	dbURL, err := h.Shortifier.ResolveShortURL(ctx, url, traceLogger)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Location", dbURL.OriginalURL)
	shared.WriteResponse(w, nil, http.StatusTemporaryRedirect)
}

func (h *Handlers) GetURLsByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "GetURLsByUserID")

	claims, err := getClaims(ctx)
	if err != nil {
		shared.WriteResponse(w, nil, http.StatusUnauthorized)
		return
	}

	urls, err := h.Shortifier.GetURLsByUserID(ctx, claims.UserID, traceLogger)
	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	var status int
	var response []byte
	if len(urls) == 0 {
		status = http.StatusNoContent
	} else {
		response, _ = json.Marshal(urls)
		status = http.StatusOK
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, response, status)
}
