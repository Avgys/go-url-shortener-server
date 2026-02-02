package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/shared"
)

const acceptEncoding = "Accept-Encoding"
const contentEncoding = "Content-Encoding"
const contentType = "Content-Type"

func WithCompression(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if isCompessableType(r.Header.Get(contentType)) {

			cBody, err := newCompressReader(r.Body, r.Header.Get(contentEncoding))

			if err != nil {
				shared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			r.Body = cBody

			w, err = newCompressWriter(w, r.Header.Get(acceptEncoding))

			if err != nil {
				shared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

var typesToCompress = []string{"application/json", "text/plain"}

func isCompessableType(contentType string) bool {

	for _, typesToCompress := range typesToCompress {
		if strings.Contains(contentType, typesToCompress) {
			return true
		}
	}
	return false
}

func newCompressReader(r io.ReadCloser, contentType string) (io.ReadCloser, error) {
	if strings.Contains(contentType, "gzip") {
		return gzip.NewReader(r)
	}

	return r, nil
}

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w compressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

const gzipType = "gzip"

func newCompressWriter(w http.ResponseWriter, contentType string) (http.ResponseWriter, error) {
	if strings.Contains(contentType, gzipType) {

		gzipWriter, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)

		if err != nil {
			return nil, err
		}

		w.Header().Set(contentEncoding, gzipType)
		w = compressWriter{ResponseWriter: w, Writer: gzipWriter}
	}

	return w, nil
}
