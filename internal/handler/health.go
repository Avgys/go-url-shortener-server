package handler

import (
	"net/http"

	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	// traceLogger := logger.Endpoint(r.Context(), "ping db")

	err := h.Store.TestConnection(r.Context())

	if err != nil {

		shared.WriteResponse(w, nil, http.StatusInternalServerError)
	}

	shared.WriteResponse(w, nil, http.StatusOK)
}
