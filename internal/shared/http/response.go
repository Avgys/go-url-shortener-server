package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/service"
)

const maxBody = 1 << 20

func WriteResponse(w http.ResponseWriter, resp []byte, code int) {
	w.WriteHeader(code)
	if len(resp) > 0 {
		w.Write(resp)
	}
}

func GetErrorStatusCode(err error) int {
	if errors.Is(err, service.ErrInvalidURL) ||
		errors.Is(err, service.ErrEmptyURL) ||
		errors.Is(err, ErrEmptyParamBody) {
		return http.StatusBadRequest
	} else if errors.Is(err, service.ErrCollision) {
		return http.StatusServiceUnavailable
	} else if errors.Is(err, service.ErrNotFound) {
		return http.StatusNotFound
	} else {
		return http.StatusInternalServerError
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {

	log.Printf("error proccessing request, %s", err.Error())

	errorText := ""

	logger.LogRequest(r, err)

	if statusCode >= 500 {
		errorText = http.StatusText(statusCode)
	} else {
		errorText = err.Error()
	}

	http.Error(w, errorText, statusCode)
}
