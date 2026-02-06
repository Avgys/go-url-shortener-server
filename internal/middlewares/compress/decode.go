package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type CompressReader struct {
	io.ReadCloser
	DecodeType string
}

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

func NewCompressReader(r *http.Request) (*CompressReader, error) {

	contentType := r.Header.Get(contentTypeHeader)
	reqDecodeType := r.Header.Get(contentEncodingHeader)

	decodeType := getDecodeType(contentType, reqDecodeType)

	if decodeType == noResult {
		return &CompressReader{r.Body, gzipType}, nil
	}

	if strings.Contains(reqDecodeType, gzipType) {
		r, err := gzip.NewReader(r.Body)
		return &CompressReader{r, gzipType}, err
	}

	return &CompressReader{r.Body, gzipType}, nil
}
