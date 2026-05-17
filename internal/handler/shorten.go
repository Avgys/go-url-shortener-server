package handler

import (
	"encoding/json"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/model/responses"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/service/shortifier"
	shared "go-url-shortener/internal/shared/http"
)

type Handlers struct {
	Shortifier *shortifier.Shortifier
	Store      repository.Repository
}

func NewHandlers(shortifier *shortifier.Shortifier, store repository.Repository) *Handlers {
	return &Handlers{Shortifier: shortifier, Store: store}
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	body, err := getBody(w, r)
	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	url := string(body)

	resultURL, err := getShortURL(h, url, traceLogger, r)

	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	var status int

	if resultURL.IsCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	w.Header().Set("Content-type", "text/plain")
	shared.WriteResponse(w, []byte(resultURL.ShortURL), status, traceLogger)
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	var reqModel requests.ShortenReq

	if err := getJSONBody(r, &reqModel); err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	resultURL, err := getShortURL(h, reqModel.URL, traceLogger, r)
	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	var status int

	if resultURL.IsCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	result, err := json.Marshal(responses.ShortenResp{URL: resultURL.ShortURL})

	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, status, traceLogger)
}

func (h *Handlers) ShortenBatch(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.FromContext(r.Context(), logger.GetFuncName())

	var reqModel requests.ShortenBatchReq

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqModel)

	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	shortenBatch, err := shortenBatch(h, reqModel, traceLogger, r)

	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	result, err := json.Marshal(shortenBatch)

	if err != nil {
		if shared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, http.StatusCreated, traceLogger)
}
