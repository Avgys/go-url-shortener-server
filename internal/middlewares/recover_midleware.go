package middlewares

import (
	"fmt"
	"go-url-shortener/internal/logger"
	"net/http"
	"os"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.FromContext(r.Context(), logger.GetFuncName())

		defer func() {
			if rvr := recover(); rvr != nil {

				traceLogger.
					Error().
					Str("recover", fmt.Sprintf("recovered in f %s", rvr)).
					Send()
				os.Exit(2)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
