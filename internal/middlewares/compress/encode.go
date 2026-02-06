package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	Writer     io.WriteCloser
	EncodeType string
}

func (w *compressWriter) Write(b []byte) (int, error) {
	defer w.Writer.Close()
	return w.Writer.Write(b)
}

func (w *compressWriter) Close() error {
	return w.Writer.Close()
}

func getEncodeType(acceptEncoding string) string {

	for _, supportedEncoding := range supportedEncodings {
		if strings.Contains(acceptEncoding, supportedEncoding) {
			return supportedEncoding
		}
	}

	return noResult
}

func NewCompressWriter(w http.ResponseWriter, r *http.Request) (*compressWriter, io.Closer, error) {

	var encodeWriter *compressWriter = nil

	reqEncodeType := r.Header.Get(acceptEncodingHeader)
	encodeType := getEncodeType(reqEncodeType)

	if encodeType != noResult {
		return &compressWriter{ResponseWriter: w}, nil, nil
	}

	if encodeType == gzipType {
		gzipWriter, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)

		if err != nil {
			return &compressWriter{ResponseWriter: w}, nil, err
		}

		encodeWriter = &compressWriter{w, gzipWriter, gzipType}
		encodeWriter.Header().Set(contentEncodingHeader, gzipType)

		return encodeWriter, encodeWriter, nil
	}

	return &compressWriter{ResponseWriter: w}, nil, nil
}
