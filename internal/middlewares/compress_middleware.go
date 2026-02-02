package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"

	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

const acceptEncodingHeader = "Accept-Encoding"
const contentEncodingHeader = "Content-Encoding"
const contentTypeHeader = "Content-Type"
const gzipType = "gzip"

var (
	errNoEncoder = errors.New("no encoder")
)

func WithCompression(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if isDecompessableType(r.Header.Get(contentTypeHeader), r.Header.Get(contentEncodingHeader)) {

			cBody, err := newCompressReader(r.Body, r.Header.Get(contentEncodingHeader))

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			r.Body = cBody
		}

		if isCompessableType(r.Header.Get(acceptEncodingHeader)) {

			var closer io.Closer = nil
			writer, closer, err := newCompressWriter(w, r.Header.Get(acceptEncodingHeader))

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			w = writer

			if closer != nil {
				defer closer.Close()
			}

		}

		h.ServeHTTP(w, r)
	})
}

var typesToDecompress = []string{"application/json", "text/plain", "application/x-gzip"}
var supportedEncodings = []string{"gzip"}

func isDecompessableType(contentType, contentEncoding string) bool {

	isSupportedContentType := false

	for _, typesToCompress := range typesToDecompress {
		if strings.Contains(contentType, typesToCompress) {
			isSupportedContentType = true
			break
		}
	}

	isSupportedEncodeType := false

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(contentEncoding, supportedEncoding) {
			isSupportedEncodeType = true
			break
		}
	}

	return isSupportedContentType && isSupportedEncodeType
}

func isCompessableType(acceptEncoding string) bool {

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(acceptEncoding, supportedEncoding) {
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
		encodeWriter.Header().Set(contentEncodingHeader, gzipType)
	}

	if encodeWriter != nil {
		return encodeWriter, encodeWriter, nil
	}

	return w, nil, nil
}
