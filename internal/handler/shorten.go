package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/model/requests"
	"github.com/Avgys/go-url-shortener-server/internal/model/responses"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Shortifier Shortifier
	Store      repository.Repository
}

type Shortifier interface {
	ResolveShortURL(ctx context.Context, model string, logerr *zerolog.Logger) (*model.DBURL, error)
	ShortifyBatch(ctx context.Context, model *service.ShortenBatchReq, logger *zerolog.Logger) (responses.ShortenBatchResp, error)
	GetURLsByUserID(ctx context.Context, userID int64, traceLogger *zerolog.Logger) ([]responses.URLPair, error)
	DeleteUrls(ctx context.Context, userID int64, urls []string, traceLogger *zerolog.Logger) error
}

func NewHandlers(shortifier Shortifier, store repository.Repository) *Handlers {
	return &Handlers{Shortifier: shortifier, Store: store}
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "ShortifyURL")

	body, err := getBody(w, r)
	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	url := string(body)

	resultURL, err := getShortURL(h, url, traceLogger, r)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
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

	ctx := r.Context()

	traceLogger := logger.Endpoint(ctx, "ShortenURL")

	var reqModel requests.ShortenReq

	if err := getJSONBody(r, &reqModel); err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	resultURL, err := getShortURL(h, reqModel.URL, traceLogger, r)
	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	var status int

	if resultURL.IsCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	result, err := json.Marshal(responses.ShortenResp{URL: resultURL.ShortURL})

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, status)
}

func (h *Handlers) ShortenBatch(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.Endpoint(r.Context(), "ShortenBatch")

	var reqModel requests.ShortenBatchReq

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqModel)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	shortenBatch, err := shortenBatch(h, reqModel, traceLogger, r)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
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
