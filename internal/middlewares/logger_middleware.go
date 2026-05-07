package middlewares

import (
	"avgys-gophermat/internal/logger"
	"fmt"
	"math/rand"
	"net/http"
	"time"
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
		log, close, err := logger.NewRequestLogger(reqCtx, spanID)

		if err != nil {
			fmt.Print("couldn't create request logger")

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer func() { _ = close() }()

		reqCtx = log.WithContext(reqCtx)
		r = r.WithContext(reqCtx)

		log, close, err = logger.Middleware(r.Context(), "WithLogging")

		if err != nil {
			fmt.Print("couldn't create request logger")

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer func() { _ = close() }()

		log.Info().
			Str("Path", r.RequestURI).
			Str("Method", r.Method).
			Str("Content-type", r.Header.Get("Content-type")).
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
