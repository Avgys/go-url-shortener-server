package handler

import (
	"encoding/json"
	"net/http"

	"go-url-shortener/internal/logger"
	dbmodel "go-url-shortener/internal/model/db"
	httphelper "go-url-shortener/internal/shared/http"
)

func (h *Handlers) GetURLsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	dbURL, err := h.Store.GetURLsStats(ctx)

	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	response := dbmodel.GetURLsStatsRow{
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
