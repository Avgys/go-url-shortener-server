package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

const acceptEncoding = "Accept-Encoding"
const contentEncoding = "Content-Encoding"
const contentType = "Content-Type"
const gzipType = "gzip"

var (
	errNoEncoder = errors.New("no encoder")
)

func WithCompression(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if isCompessableType(r.Header.Get(contentType)) {

			cBody, err := newCompressReader(r.Body, r.Header.Get(contentEncoding))

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			r.Body = cBody
			var closer io.Closer = nil

			w, closer, err = newCompressWriter(w, r.Header.Get(acceptEncoding))

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			if closer != nil {
				defer closer.Close()
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
	Writer io.WriteCloser
}

func (w *compressWriter) Write(b []byte) (int, error) {
	defer w.Writer.Close()
	return w.Writer.Write(b)
}

func (w *compressWriter) Close() error {
	return w.Writer.Close()
}

func newCompressWriter(w http.ResponseWriter, contentType string) (http.ResponseWriter, io.Closer, error) {

	var encodeWriter *compressWriter = nil

	if strings.Contains(contentType, gzipType) {
		gzipWriter, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)

		if err != nil {
			return nil, nil, err
		}

		encodeWriter = &compressWriter{ResponseWriter: w, Writer: gzipWriter}
		encodeWriter.Header().Set(contentEncoding, gzipType)
	}

	if encodeWriter != nil {
		return encodeWriter, encodeWriter, nil
	}

	return w, nil, nil
}
