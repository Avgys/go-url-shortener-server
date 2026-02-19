package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Shortifier Shortifier
	Store      repository.Repository
}

type Shortifier interface {
	ResolveShortURL(ctx context.Context, model string, logerr *zerolog.Logger) (string, error)
	ShortifyURL(ctx context.Context, model string, logger *zerolog.Logger) (*model.IndexedShortURL, error)
	ShortifyBatch(ctx context.Context, model []model.IndexedFullURL, logger *zerolog.Logger) (model.ShortenBatchResp, error)
}

func NewHandlers(shortifier Shortifier, store repository.Repository) *Handlers {
	return &Handlers{Shortifier: shortifier, Store: store}
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.Endpoint(r.Context(), "ShortifyURL")

	body, shouldReturn := getBody(w, r, traceLogger)
	if shouldReturn {
		return
	}

	url := string(body)

	resultURL, shouldReturn := getShortURL(h, url, traceLogger, w, r)
	if shouldReturn {
		return
	}

	var status int

	if resultURL.IsCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	w.Header().Set("Content-type", "text/plain")
	shared.WriteResponse(w, []byte(resultURL.ShortURL), status)
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.Endpoint(r.Context(), "ShortenURL")

	var reqModel model.ShortenReq

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqModel)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	resultURL, shouldReturn := getShortURL(h, reqModel.URL, traceLogger, w, r)
	if shouldReturn {
		return
	}

	var status int

	if resultURL.IsCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	result, err := json.Marshal(model.ShortenResp{URL: resultURL.ShortURL})

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, status)
}

func (h *Handlers) ShortenBatch(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.Endpoint(r.Context(), "ShortenBatch")

	var reqModel model.ShortenBatchReq

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqModel)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	shortenBatch, shouldReturn := getShortBatch(h, reqModel, traceLogger, w, r)
	if shouldReturn {
		return
	}

	result, err := json.Marshal(shortenBatch)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, http.StatusCreated)
}
