package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	EncodeWriter io.WriteCloser
	EncodeType   string
}

func (w *compressWriter) Write(b []byte) (int, error) {
	if w.EncodeWriter != nil {
		return w.EncodeWriter.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func (w *compressWriter) Close() error {
	if w.EncodeWriter != nil {
		return w.EncodeWriter.Close()
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

	var encodeWriter *compressWriter = nil

	reqEncodeType := r.Header.Get(acceptEncodingHeader)
	encodeType := getEncodeType(reqEncodeType)

	if encodeType != noResult {
		return &compressWriter{ResponseWriter: w}, nil
	}

	if encodeType == gzipType {
		gzipWriter, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)

		if err != nil {
			return &compressWriter{ResponseWriter: w}, err
		}

		encodeWriter = &compressWriter{w, gzipWriter, gzipType}
		encodeWriter.Header().Set(contentEncodingHeader, gzipType)

		return encodeWriter, nil
	}

	return &compressWriter{ResponseWriter: w}, nil
}
