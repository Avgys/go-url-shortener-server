package handler

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

func (h *Handlers) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "DeleteShortURL")

	claims, err := getClaims(ctx)
	if err != nil {
		shared.WriteResponse(w, nil, http.StatusUnauthorized)
		return
	}

	var urls []string

	if err = getJSONBody(r, &urls); err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	err = h.Shortifier.DeleteUrls(ctx, claims.UserID, urls, traceLogger)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	shared.WriteResponse(w, nil, http.StatusAccepted)
}
