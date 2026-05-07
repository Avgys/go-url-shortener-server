package handler

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/auth/jwttoken"
	"github.com/Avgys/go-url-shortener-server/internal/logger"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

func (h *Handlers) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "DeleteShortURL")

	claims, err := jwttoken.GetClaims(ctx)

	if err != nil {
		traceLogger.Err(err).Send()
		err = shared.NewError("unauthorized/broken token", http.StatusUnauthorized)
		shared.WriteError(w, r, err, traceLogger)
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
