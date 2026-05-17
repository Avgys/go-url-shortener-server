package middlewares

import (
	"fmt"
	"go-url-shortener/internal/logger"
	"net/http"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {

		traceLogger := logger.FromContext(req.Context(), logger.GetFuncName())

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
	})
}
