package middlewares

import (
	"fmt"
	"net/http"

	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/middlewares/compress"
	httpShared "go-url-shortener/internal/shared/http"
)

func WithCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger, close, err := logger.Middleware(r.Context(), "compress")

		if err != nil {
			fmt.Print("couldn't create request logger")

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer func() { _ = close() }()

		decodeReader, err := compress.NewCompressReader(r)

		if err != nil {
			httpShared.HandleErr(w, r, err, traceLogger)
			return
		}

		r.Body = decodeReader

		encodeWriter, err := compress.NewCompressWriter(w, r)

		if err != nil {
			httpShared.HandleErr(w, r, err, traceLogger)
			return
		}

		defer func() { _ = encodeWriter.Close() }()

		w = encodeWriter

		traceLogger.Info().
			Str("Request Content-type", r.Header.Get("Content-Type")).
			Str("Request Decode-type", r.Header.Get("Content-Encoding")).
			Str("Decode-type", decodeReader.DecodeType).
			Str("Request Encode-type", r.Header.Get("Accept-Encoding")).
			Str("Encode-type", encodeWriter.EncodeType).
			Send()

		h.ServeHTTP(w, r)
	})
}
