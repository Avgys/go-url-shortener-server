package handler

import (
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
	"github.com/go-chi/chi/v5"
)

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	url, err := getRequestBody(r)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	isCreated := false
	resultURL := ""

	if resultURL, isCreated, err = h.Shortifier.ShortifyURL(url); err != nil {
		if errors.Is(err, shortifier.ErrInvalidURL) {
			writeError(w, r, err, http.StatusBadRequest)
		} else {
			writeError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	var status int
	if isCreated {
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}

	writeResponse(w, resultURL, status)
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	url = chi.URLParam(r, "url")

	if url, err = h.Shortifier.ResolveShortURL(url); err != nil {
		writeError(w, r, err, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)

	writeResponse(w, "", http.StatusTemporaryRedirect)
}
