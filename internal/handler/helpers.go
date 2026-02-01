package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/service"
)

const maxBody = 1 << 20

var (
	errInternalErrorReadingBody = errors.New("got error reading url")
	ErrEmptyParamBody           = errors.New("empty param body")
)

func getRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {

	buffer := make([]byte, 128)
	readBytes := 0
	var err error

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	if readBytes, err = r.Body.Read(buffer); err != nil && err != io.EOF {
		fmt.Printf("got error %v\n", err)

		return nil, errInternalErrorReadingBody
	}

	if readBytes == 0 {
		return nil, ErrEmptyParamBody
	}

	return buffer[:readBytes], nil
}

func writeResponse(w http.ResponseWriter, resp []byte, code int) {

	w.WriteHeader(code)
	if len(resp) > 0 {
		w.Write(resp)
	}
}

func getErrorStatusCode(err error) int {
	if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrEmptyURL) {
		return http.StatusBadRequest
	} else if errors.Is(err, service.ErrCollision) {
		return http.StatusServiceUnavailable
	} else if errors.Is(err, service.ErrNotFound) {
		return http.StatusNotFound
	} else {
		return http.StatusInternalServerError
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {

	log.Printf("error proccessing request, %s", err.Error())

	errorText := ""

	logRequest(r, err)

	if statusCode >= 500 {
		errorText = http.StatusText(statusCode)
	} else {
		errorText = err.Error()
	}

	http.Error(w, errorText, statusCode)
}

func logRequest(r *http.Request, err error) {
	payload := struct {
		Method      string      `json:"method"`
		Path        string      `json:"path"`
		Query       string      `json:"query"`
		Headers     http.Header `json:"headers"`
		ContentType string      `json:"contentType"`
		RemoteAddr  string      `json:"remoteAddr"`
		Error       string      `json:"error"`
	}{
		Method:      r.Method,
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
		Headers:     r.Header,
		ContentType: r.Header.Get("Content-Type"),
		RemoteAddr:  r.RemoteAddr,
		Error:       err.Error(),
	}

	enc := json.NewEncoder(log.Writer())
	enc.SetIndent("", "  ")
	encErr := enc.Encode(payload)

	if encErr != nil {
		logger.Log.
			Error().
			Str("error_reason", "error marshaling request payload").
			Err(encErr)
	}
}
