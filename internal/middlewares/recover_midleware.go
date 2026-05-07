package middlewares

import (
	"avgys-gophermat/internal/logger"
	"fmt"
	"net/http"
)

func Recoverer(next http.Handler) http.Handler {

	fn := func(writer http.ResponseWriter, req *http.Request) {

		traceLogger := logger.FromContext(req.Context())

		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				traceLogger.
					Error().
					Str("recover", fmt.Sprintf("recovered in f %s", rvr)).
					Send()
			}
		}()

		next.ServeHTTP(writer, req)
	}

	return http.HandlerFunc(fn)
}
