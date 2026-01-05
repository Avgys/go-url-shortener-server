package handler

import (
	"errors"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/service/shortifier"
)

var (
	ErrWrongContentType         = errors.New("wrong content-type")
	ErrInternalErrorReadingBody = errors.New("got error reading url")
	ErrEmptyParamBody           = errors.New("empty param body")
)

type Handlers struct {
	Shortifier Shortifier
}

func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	if url, err = getURIParam(r); err != nil {
		var statusCode int

		if errors.Is(err, ErrInternalErrorReadingBody) {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusBadRequest
		}

		writeError(w, r, err, statusCode)
		return
	}

	if url, err = h.Shortifier.ResolveShortURL(url); err != nil {
		writeError(w, r, err, http.StatusNotFound)
		return
	}

	w.Header().Set("Location", url)

	writeResponse(w, "", http.StatusTemporaryRedirect)
}

func (h *Handlers) ShortifyURL(w http.ResponseWriter, r *http.Request) {

	url, err := getRequestBody(r)
	if err != nil {
		writeError(w, r, err, http.StatusBadRequest)
		return
	}

	isCreated := false

	if url, isCreated, err = h.Shortifier.ShortifyURL(url); err != nil {
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

	writeResponse(w, url, status)
}
