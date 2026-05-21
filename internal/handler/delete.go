package handler

import (
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
	httpshared "go-url-shortener/internal/shared/http"
)

func (h *Handlers) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	claims, err := auth.GetFromContext(ctx)

	if err != nil {
		traceLogger.Err(err).Send()
		err = httpshared.NewError("unauthorized/broken token", http.StatusUnauthorized)

		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	var urls []string

	err = httpshared.GetJSONBody(r, &urls)

	if httpshared.HandleErr(w, r, err, traceLogger) {
		return
	}

	err = h.Shortifier.DeleteUrls(ctx, claims.UserID, urls, traceLogger)

	if httpshared.HandleErr(w, r, err, traceLogger) {
		return
	}

	httpshared.WriteResponse(w, nil, http.StatusAccepted, traceLogger)
}
