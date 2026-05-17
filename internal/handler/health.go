package handler

import (
	"net/http"

	"go-url-shortener/internal/logger"
	httphelper "go-url-shortener/internal/shared/http"
)

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	err := h.Store.TestConnection(r.Context())

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	httphelper.WriteResponse(w, nil, http.StatusOK, traceLogger)
}
