package handler

import (
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/service/auth"
	httphelper "go-url-shortener/internal/shared/http"
)

func (h *Handlers) DeleteShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	claims, err := auth.GetFromContext(ctx)

	if err != nil {
		traceLogger.Err(err).Send()
		err = httphelper.NewError("unauthorized/broken token", http.StatusUnauthorized)

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

	httphelper.WriteResponse(w, nil, http.StatusAccepted, traceLogger)
}
