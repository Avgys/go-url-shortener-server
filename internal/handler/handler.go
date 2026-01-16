package handler

import (
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Shortifier Shortifier
}

type Shortifier interface {
	ResolveShortURL(shortURL string) (string, error)
	ShortifyURL(url string) (string, error)
}

func NewHandlers(shortifier Shortifier) *Handlers {
	return &Handlers{Shortifier: shortifier}
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	url, err := getRequestBody(r)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}
	resultURL := ""

	if resultURL, err = h.Shortifier.ShortifyURL(url); err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			writeError(w, r, err, http.StatusBadRequest)
		} else {
			writeError(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	writeResponse(w, resultURL, http.StatusCreated)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(url); err != nil {

		if errors.Is(err, service.ErrNotFound) {
			writeError(w, r, err, http.StatusNotFound)
		} else {
			writeError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)

	writeResponse(w, "", http.StatusTemporaryRedirect)
}
