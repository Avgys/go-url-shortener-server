package handler

import (
	"encoding/json"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/model/responses"
	httphelper "go-url-shortener/internal/shared/http"
)

func (h *Handlers) GetURLsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	dbURL, err := h.Store.GetURLsStats(ctx)

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	response := responses.GetURLsStatsResponse{
		UniqueLongUrlCount: dbURL.UniqueLongUrlCount,
		UniqueUserIDCount:  dbURL.UniqueUserIDCount,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		httphelper.HandleErr(w, r, err, traceLogger)
		return
	}

	httphelper.WriteResponse(w, responseBytes, http.StatusOK, traceLogger)
}
