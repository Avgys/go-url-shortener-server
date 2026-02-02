package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/model"
	shared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
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

	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	url := string(body)
	resultURL := ""

	if resultURL, err = h.Shortifier.ShortifyURL(url); err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	w.Header().Set("Content-type", "text/plain")
	shared.WriteResponse(w, []byte(resultURL), http.StatusCreated)
}

func (h *Handlers) ShortenURL(w http.ResponseWriter, r *http.Request) {

	body, err := shared.GetRequestBody(w, r)

	if err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	var reqModel model.ShortenReq

	err = json.Unmarshal(body, &reqModel)

	if err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	resultURL, err := h.Shortifier.ShortifyURL(reqModel.URL)

	if err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	result, err := json.Marshal(model.ShortenResp{URL: resultURL})

	if err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	w.Header().Set("Content-type", "application/json")
	shared.WriteResponse(w, result, http.StatusCreated)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(url); err != nil {
		shared.WriteError(w, r, err, shared.GetErrorStatusCode(err))
		return
	}

	w.Header().Set("Location", url)
	shared.WriteResponse(w, nil, http.StatusTemporaryRedirect)
}
