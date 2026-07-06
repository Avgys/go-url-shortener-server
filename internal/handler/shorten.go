package handler

import (
	"encoding/json"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/model/requests"
	"go-url-shortener/internal/model/responses"
	httpshared "go-url-shortener/internal/shared/http"
)

// ShortifyURL handles POST / with a plain-text body containing the long URL.
// On success it responds with 201 Created or 409 Conflict when the URL already exists,
// and writes the short URL as text/plain.
func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	body, err := httpshared.GetRequestBody(w, r)
	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	url := string(body)

	resultURL, err := h.Shortifier.ShortenURL(ctx, url, traceLogger)

	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
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
	httpshared.WriteResponseStr(w, resultURL.ShortURL, status, traceLogger)
}

// ShortenURL handles POST /api/shorten with a JSON body {"url":"..."}.
// The response is application/json with a result field holding the short URL.
func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	traceLogger := logger.FromContext(ctx, logger.GetFuncName())

	var reqModel requests.ShortenReq

	if err := httpshared.GetJSONBody(r, &reqModel); err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	resultURL, err := h.Shortifier.ShortenURL(ctx, reqModel.URL, traceLogger)
	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
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
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	httpshared.WriteResponse(w, result, status, traceLogger)
}

// ShortenBatch handles POST /api/shorten/batch with a JSON array of correlation_id and url pairs.
// It responds with 201 Created and a JSON array of shortened URLs.
func (h *Handlers) ShortenBatch(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.FromContext(r.Context(), logger.GetFuncName())

	var reqModel requests.ShortenBatchReq

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqModel)

	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	shortenBatch, err := h.Shortifier.ShortenBatch(r.Context(), reqModel, traceLogger)

	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	result, err := json.Marshal(shortenBatch)

	if err != nil {
		if httpshared.HandleErr(w, r, err, traceLogger) {
			return
		}
	}

	w.Header().Set("Content-type", "application/json")
	httpshared.WriteResponse(w, result, http.StatusCreated, traceLogger)
}
