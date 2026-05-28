package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	baseWriter http.ResponseWriter
	buffer     bytes.Buffer

	EncodeType string

	codeStatus       int
	compressRequired bool
}

func (w *compressWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

func (w *compressWriter) Header() http.Header {
	return w.baseWriter.Header()
}

func (w *compressWriter) WriteHeader(statusCode int) {
	w.codeStatus = statusCode
}

func (w *compressWriter) Close() error {
	return w.flush()
}

func (w *compressWriter) flush() error {
	if w.buffer.Len() > minGzipResponseBytes {
		w.compressRequired = true
	}

	var resultWriter io.Writer = w.baseWriter
	var closeWriter io.Closer = nil

	if w.compressRequired {
		if w.EncodeType == gzipType {
			gzipWriter, err := gzip.NewWriterLevel(w.baseWriter, gzip.DefaultCompression)

			if err != nil {
				return err
			}

			resultWriter = gzipWriter
			closeWriter = gzipWriter
			w.baseWriter.Header().Set(contentEncodingHeader, gzipType)
		}
	}

	resultWriter.Write(w.buffer.Bytes())
	w.baseWriter.WriteHeader(w.codeStatus)

	if closeWriter != nil {
		return closeWriter.Close()
	}

	return nil
}

func getEncodeType(acceptEncoding string) string {

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(acceptEncoding, supportedEncoding) {
			return supportedEncoding
		}
	}

	return noResult
}

func NewCompressWriter(w http.ResponseWriter, r *http.Request) (*compressWriter, error) {

	reqEncodeType := r.Header.Get(acceptEncodingHeader)
	encodeType := getEncodeType(reqEncodeType)

	if encodeType == gzipType {
		return &compressWriter{baseWriter: w, EncodeType: encodeType}, nil
	}

	return &compressWriter{baseWriter: w}, nil
}
