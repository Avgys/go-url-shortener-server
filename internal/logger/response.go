package logger

import (
	"encoding/json"
	"log"
	"net/http"
)

func LogRequest(r *http.Request, err error) {
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

	traceLogger := FromContext(r.Context())

	if encErr != nil {
		traceLogger.
			Error().
			Str("error_reason", "error marshaling request payload").
			Err(encErr)
	}
}
