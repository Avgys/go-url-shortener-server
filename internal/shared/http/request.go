package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
)

func GetRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {

	buffer := make([]byte, 128)
	readBytes := 0
	var err error

	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	if readBytes, err = r.Body.Read(buffer); err != nil && err != io.EOF {
		fmt.Printf("got error %v\n", err)

		return nil, fmt.Errorf("got error reading url: %w", err)
	}

	if readBytes == 0 {
		return nil, logger.NewError("empty param body", http.StatusBadRequest)
	}

	result := buffer[:readBytes]

	traceLogger := logger.FromContext(r.Context())
	traceLogger.Info().
		Str("Request body", string(result))

	return result, nil
}
