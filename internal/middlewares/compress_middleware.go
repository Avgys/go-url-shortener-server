package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/logger"
	httpShared "github.com/Avgys/go-url-shortener-server/internal/shared/http"
)

const acceptEncodingHeader = "Accept-Encoding"
const contentEncodingHeader = "Content-Encoding"
const contentTypeHeader = "Content-Type"
const gzipType = "gzip"
const noResult = "no result"

func WithCompression(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		traceLogger := logger.FromContext(r.Context()).With().Str("middleware", "compress").Logger()

		reqContentType := r.Header.Get(contentTypeHeader)
		reqDecodeType := r.Header.Get(contentEncodingHeader)

		decodeType := getDecodeType(reqContentType, reqDecodeType)
		if decodeType != noResult {

			cBody, err := newCompressReader(r.Body, decodeType)

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			r.Body = cBody
		}

		reqEncodeType := r.Header.Get(acceptEncodingHeader)
		encodeType := getEncodeType(reqEncodeType)

		if encodeType != noResult {

			var closer io.Closer = nil
			writer, closer, err := newCompressWriter(w, encodeType)

			if err != nil {
				httpShared.WriteError(w, r, err, http.StatusInternalServerError)
				return
			}

			w = writer

			if closer != nil {
				defer closer.Close()
			}

		}

		traceLogger.Info().
			Str("Request Content-type", reqContentType).
			Str("Request Decode-type", reqDecodeType).
			Str("Decode-type", decodeType).
			Str("Request Encode-type", reqEncodeType).
			Str("Encode-type", encodeType).
			Send()

		h.ServeHTTP(w, r)
	})
}

var typesToDecompress = []string{"application/json", "text/plain", "application/x-gzip"}
var supportedEncodings = []string{gzipType}

func getDecodeType(contentType, contentEncoding string) string {

	isSupportedContentType := false

	for _, typesToCompress := range typesToDecompress {
		if strings.Contains(contentType, typesToCompress) {
			isSupportedContentType = true
			break
		}
	}

	if !isSupportedContentType {
		return noResult
	}

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(contentEncoding, supportedEncoding) {
			return supportedEncoding
		}
	}

	return noResult
}

func getEncodeType(acceptEncoding string) string {

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(acceptEncoding, supportedEncoding) {
			return supportedEncoding
		}
	}

	return noResult
}

func newCompressReader(r io.ReadCloser, contentType string) (io.ReadCloser, error) {
	if strings.Contains(contentType, gzipType) {
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

	if contentType == gzipType {
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
