package httphelper

import (
	"avgys-gophermat/internal/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func GetJSONBody(r *http.Request, value any) error {

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(value)

	if err != nil {
		var syntaxErr *json.SyntaxError

		if errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &syntaxErr) {
			err = NewError(err.Error(), http.StatusBadRequest)
		}
	}

	return err
}

func GetRequestBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {

	traceLogger := logger.FromContext(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	result, err := io.ReadAll(r.Body)
	if err != nil {
		traceLogger.Err(err).Msg("got error reading body")

		return nil, fmt.Errorf("got error reading body: %w", err)
	}

	if len(result) == 0 {
		return nil, NewError("empty param body", http.StatusBadRequest)
	}

	traceLogger.Info().
		Str("Request body", string(result)).
		Send()

	return result, nil
}
