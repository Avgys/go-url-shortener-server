package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

var (
	errInternalErrorReadingBody = errors.New("got error reading url")
	ErrEmptyParamBody           = errors.New("empty param body")
)

func GetRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {

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

	result := buffer[:readBytes]

	traceLogger := logger.FromContext(r.Context())
	traceLogger.Info().
		Str("Request body", string(result))

	return result, nil
}
