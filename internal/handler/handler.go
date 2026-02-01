package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Shortifier Shortifier
}

type Shortifier interface {
	ResolveShortURL(model string) (string, error)
	ShortifyURL(model string) (string, error)
}

func NewHandlers(shortifier Shortifier) *Handlers {
	return &Handlers{Shortifier: shortifier}
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	body, err := getRequestBody(w, r)

	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	url := string(body)
	resultURL := ""

	if resultURL, err = h.Shortifier.ShortifyURL(url); err != nil {
		writeError(w, r, err, getErrorStatusCode(err))
		return
	}

	writeResponse(w, []byte(resultURL), http.StatusCreated)
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	body, err := getRequestBody(w, r)

	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	var reqModel model.ShortenReq

	err = json.Unmarshal(body, &reqModel)

	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	resultURL, err := h.Shortifier.ShortifyURL(reqModel.Url)

	if err != nil {
		writeError(w, r, err, getErrorStatusCode(err))
		return
	}

	result, err := json.Marshal(model.ShortenResp{Url: resultURL})

	if err != nil {
		writeError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	writeResponse(w, result, http.StatusCreated)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(url); err != nil {

		writeError(w, r, err, getErrorStatusCode(service.ErrNotFound))
		return
	}

	w.Header().Set("Location", url)
	writeResponse(w, nil, http.StatusTemporaryRedirect)
}
