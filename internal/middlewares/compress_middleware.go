package middlewares

import (
	"net/http"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	"github.com/Avgys/go-url-shortener-server/internal/middlewares/compress"
	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

func WithCompression(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.Middleware(r.Context(), "compress")

		decodeReader, err := compress.NewCompressReader(r)

		if err != nil {
			httpShared.WriteError(w, r, err, traceLogger)
			return
		}

		r.Body = decodeReader

		encodeWriter, closer, err := compress.NewCompressWriter(w, r)

		if err != nil {
			httpShared.WriteError(w, r, err, traceLogger)
			return
		}

		if closer != nil {
			defer closer.Close()
		}

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
