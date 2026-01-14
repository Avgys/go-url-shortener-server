package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	errWrongContentType         = errors.New("wrong content-type")
	errInternalErrorReadingBody = errors.New("got error reading url")
	ErrEmptyParamBody           = errors.New("empty param body")
)

func getRequestBody(r *http.Request) (string, error) {

	if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		return "", fmt.Errorf("%w: %v", errWrongContentType, r.Header.Get("Content-type"))
	}

	buffer := make([]byte, 128)
	readBytes := 0
	var err error

	if readBytes, err = r.Body.Read(buffer); err != nil && err != io.EOF {
		fmt.Printf("got error %v\n", err)

		return "", errInternalErrorReadingBody
	}

	if readBytes == 0 {
		return "", ErrEmptyParamBody
	}

	url := string(buffer[:readBytes])

	return url, nil
}

func writeResponse(w http.ResponseWriter, text string, code int) {

	w.WriteHeader(code)
	if text != "" {
		w.Write([]byte(text))
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
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

	log.Printf("error proccessing request, %s", err.Error())

	enc := json.NewEncoder(log.Writer())
	enc.SetIndent("", "  ")
	err = enc.Encode(payload)

	if err != nil {
		log.Printf("error marshaling request payload, %s", err.Error())
	}

	http.Error(w, http.StatusText(statusCode), statusCode)
}
