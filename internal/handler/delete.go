package handler

import (
	"net/http"

	"go-url-shortener/internal/auth/jwttoken"
	"go-url-shortener/internal/logger"
	httphelper "go-url-shortener/internal/shared/http"
	shared "go-url-shortener/internal/shared/http"
)

func (h *Handlers) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	claims, err := jwttoken.GetClaims(ctx)

	if err != nil {
		traceLogger.Err(err).Send()
		err = shared.NewError("unauthorized/broken token", http.StatusUnauthorized)

		if httphelper.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	var urls []string

	err = getJSONBody(r, &urls)

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	err = h.Shortifier.DeleteUrls(ctx, claims.UserID, urls, traceLogger)

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	shared.WriteResponse(w, nil, http.StatusAccepted, traceLogger)
}
