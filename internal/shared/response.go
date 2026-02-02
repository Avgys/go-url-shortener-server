package shared

import (
	"errors"
	"log"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/shared/common_errors"
)

const maxBody = 1 << 20

func WriteResponse(w http.ResponseWriter, resp []byte, code int) {
	w.WriteHeader(code)
	if len(resp) > 0 {
		w.Write(resp)
	}
}

func GetErrorStatusCode(err error) int {
	if errors.Is(err, common_errors.ErrInvalidURL) || errors.Is(err, common_errors.ErrEmptyURL) {
		return http.StatusBadRequest
	} else if errors.Is(err, common_errors.ErrCollision) {
		return http.StatusServiceUnavailable
	} else if errors.Is(err, common_errors.ErrNotFound) {
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
