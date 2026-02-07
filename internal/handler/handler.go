package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type Handlers struct {
	Shortifier Shortifier
}

type Shortifier interface {
	ResolveShortURL(model string, logerr *zerolog.Logger) (string, error)
	ShortifyURL(model string, logger *zerolog.Logger) (string, error)
}

func NewHandlers(shortifier Shortifier) *Handlers {
	return &Handlers{Shortifier: shortifier}
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

	w.Header().Set("Content-type", "text/plain")
	shared.WriteResponse(w, []byte(resultURL), http.StatusCreated)
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	traceLogger := logger.Endpoint(r.Context(), "ShortenURL")

	body, shouldReturn := getBody(w, r, traceLogger)
	if shouldReturn {
		return
	}

	var reqModel model.ShortenReq

	err := json.Unmarshal(body, &reqModel)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	resultURL, shouldReturn := getShortURL(h, reqModel.URL, traceLogger, w, r)
	if shouldReturn {
		return
	}

	result, err := json.Marshal(model.ShortenResp{URL: resultURL})

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, http.StatusCreated)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	traceLogger := logger.Endpoint(r.Context(), "Redirect")

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(url, traceLogger); err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return
	}

	w.Header().Set("Location", url)
	shared.WriteResponse(w, nil, http.StatusTemporaryRedirect)
}

func getBody(w http.ResponseWriter, r *http.Request, traceLogger *zerolog.Logger) ([]byte, bool) {
	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		shared.WriteError(w, r, err, traceLogger)
		return nil, true
	}
	return body, false
}

func getShortURL(h *Handlers, url string, traceLogger *zerolog.Logger, w http.ResponseWriter, r *http.Request) (string, bool) {
	resultURL, err := h.Shortifier.ShortifyURL(url, traceLogger)

	if err != nil {
		if errors.Is(err, service.ErrCollision) {
			err = logger.NewError(err.Error(), http.StatusTooManyRequests)
		}

		shared.WriteError(w, r, err, traceLogger)

		return "", true
	}

	return resultURL, false
}
