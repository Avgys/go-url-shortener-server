package middlewares

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/samber/lo"
)

type (
	responseLogData struct {
		statusCode   int
		responseSize int
	}

	writerWrapper struct {
		logData     responseLogData
		innerWriter http.ResponseWriter
	}
)

func WithLogging(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		spanID := rand.Int63()
		reqCtx := r.Context()
		log, close := logger.NewRequestLogger(reqCtx, spanID)

		defer close()

		reqCtx = log.WithContext(reqCtx)
		r = r.WithContext(reqCtx)

		log = logger.Middleware(r.Context(), "WithLogging")

		log.Info().
			Str("Path", r.RequestURI).
			Str("Method", r.Method).
			Str("Content-type", r.Header.Get("Content-type")).
			Str("Cookies", formatCookies(r.Cookies())).
			Msg("Started processing")

		wrappedWriter := wrapWriter(w)

		h.ServeHTTP(wrappedWriter, r)

		executionTime := time.Since(startTime)

		log.Info().
			Str("Path", r.RequestURI).
			Str("Method", r.Method).
			Dur("Excecution time", executionTime).
			Int("Response size", wrappedWriter.logData.responseSize).
			Int("Response status code", wrappedWriter.logData.statusCode).
			Msg("Request processed")
	})
}

func wrapWriter(w http.ResponseWriter) *writerWrapper {
	return &writerWrapper{
		logData:     responseLogData{},
		innerWriter: w,
	}
}

func (wr *writerWrapper) WriteHeader(statusCode int) {
	wr.innerWriter.WriteHeader(statusCode)
	wr.logData.statusCode = statusCode
}

func (wr *writerWrapper) Header() http.Header {
	return wr.innerWriter.Header()
}

func (wr *writerWrapper) Write(input []byte) (int, error) {

	size, err := wr.innerWriter.Write(input)
	wr.logData.responseSize = size

	return size, err
}

func formatCookies(cookies []*http.Cookie) string {
	return strings.Join(lo.Map(cookies, func(cookie *http.Cookie, _ int) string { return fmt.Sprintf("%s:%s", cookie.Name, cookie.Value) }), ",")
}
