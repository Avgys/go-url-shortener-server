package handler

import (
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

func getBody(w http.ResponseWriter, r *http.Request, traceLogger *zerolog.Logger) ([]byte, bool) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}

	return body, false
}

func getShortURL(h *Handlers, url string, traceLogger *zerolog.Logger, w http.ResponseWriter, r *http.Request) (*model.IndexedShortURL, bool) {
	resultURL, err := h.Shortifier.ShortifyURL(r.Context(), url, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}

	return resultURL, false
}

func getShortBatch(h *Handlers, model model.ShortenBatchReq, traceLogger *zerolog.Logger, w http.ResponseWriter, r *http.Request) (model.ShortenBatchResp, bool) {
	resultURL, err := h.Shortifier.ShortifyBatch(r.Context(), model, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = shared.NewError(err.Error(), http.StatusTooManyRequests)
		}

		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}

	return resultURL, false
}
