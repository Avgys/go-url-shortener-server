package httphelper

import (
	"encoding/json"
	"errors"

	"net/http"

	"github.com/rs/zerolog"
)

const maxBody = 1 << 20

// WriteResponseStr writes resp with the given status code and logs the response.
func WriteResponseStr(w http.ResponseWriter, resp string, code int, tracelog *zerolog.Logger) {
	// body := unsafe.Slice(unsafe.StringData(resp), len(resp))
	body := []byte(resp)
	WriteResponse(w, body, code, tracelog)
}

// WriteResponse writes resp with the given status code and logs the response.
func WriteResponse(w http.ResponseWriter, resp []byte, code int, tracelog *zerolog.Logger) {
	w.WriteHeader(code)

	tracelog.Info().Int("status code", code).Str("response", string(resp)).Msg("write response")

	if len(resp) > 0 {
		if _, err := w.Write(resp); err != nil {
			tracelog.Err(err).Msg("write response body")
		}
	}
}

// HandleErr writes err to w when non-nil. [ShowHTTPError] values set the status code;
// server errors (5xx) also log the full request. Returns true when a response was written.
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
