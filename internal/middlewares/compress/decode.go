package compress

import (
	"compress/gzip"
	"errors"
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

	reader := r.Body

	if decodeType == noResult {
		return &CompressReader{r.Body, gzipType}, nil
	}

	if strings.Contains(reqDecodeType, gzipType) {
		var err error
		reader, err = gzip.NewReader(r.Body)

		if err != nil {
			if errors.Is(err, io.EOF) {
				return &CompressReader{r.Body, contentType}, nil
			}

			return &CompressReader{r.Body, gzipType}, err
		}
	}

	return &CompressReader{reader, gzipType}, nil
}
