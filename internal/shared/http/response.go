package httphelper

import (
	"encoding/json"
	"errors"

	"net/http"

	"github.com/rs/zerolog"
)

const maxBody = 1 << 20

func WriteResponse(w http.ResponseWriter, resp []byte, code int, tracelog *zerolog.Logger) {
	w.WriteHeader(code)

	tracelog.Info().Int("status code", code).Str("response", string(resp)).Msg("write response")

	if len(resp) > 0 {
		if _, err := w.Write(resp); err != nil {
			tracelog.Err(err).Msg("write response body")
		}
	}
}

func HandleErr(w http.ResponseWriter, r *http.Request, err error, tracelog *zerolog.Logger) bool {

	if err == nil {
		return false
	}

	errorText := http.StatusText(http.StatusInternalServerError)
	statusCode := http.StatusInternalServerError

	var loggerError *ShowHTTPError
	if errors.As(err, &loggerError) {
		errorText = loggerError.Error()
		statusCode = loggerError.StatusCode
	}

	const internalError = 500
	if statusCode >= internalError {
		logRequest(r, err, tracelog)
	}

	tracelog.Info().Int("status code", statusCode).Str("response", errorText).Msg("write response")
	http.Error(w, errorText, statusCode)

	return true
}

func logRequest(r *http.Request, err error, tracelog *zerolog.Logger) {
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

	reqJSON, encErr := json.Marshal(payload)

	if encErr == nil {
		tracelog.Error().
			Err(err).
			RawJSON("Request", reqJSON).
			Msg("error processing request")
	} else {
		tracelog.
			Error().
			Err(encErr).
			Msg("error marshaling request payload")
	}
}
