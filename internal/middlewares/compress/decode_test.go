package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestCompressSuite(t *testing.T) {
	suite.Run(t, new(CompressSuite))
}

type CompressSuite struct {
	suite.Suite
}

func (s *CompressSuite) gzipBytes(data string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write([]byte(data))
	s.Require().NoError(err)
	s.Require().NoError(gz.Close())

	return buf.Bytes()
}

func (s *CompressSuite) TestGetDecodeType() {
	tests := []struct {
		name            string
		contentType     string
		contentEncoding string
		want            string
	}{
		{
			name:            "gzip with supported type",
			contentType:     "application/json",
			contentEncoding: "gzip",
			want:            gzipType,
		},
		{
			name:            "unsupported content type",
			contentType:     "text/html",
			contentEncoding: "gzip",
			want:            noResult,
		},
		{
			name:            "unsupported encoding",
			contentType:     "application/json",
			contentEncoding: "br",
			want:            noResult,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			got := getDecodeType(test.contentType, test.contentEncoding)
			s.Equal(test.want, got)
		})
	}
}

func (s *CompressSuite) TestNewCompressReader_Gzip() {
	body := s.gzipBytes("hello")
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(body))
	req.Header.Set(contentTypeHeader, "application/json")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(reader.Close())
	}()

	decompressed, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal("hello", string(decompressed))
	s.Equal(gzipType, reader.DecodeType)
}

func (s *CompressSuite) TestNewCompressReader_UnsupportedContentType() {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("plain")))
	req.Header.Set(contentTypeHeader, "text/html")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(reader.Close())
	}()

	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal("plain", string(data))
	s.Equal(gzipType, reader.DecodeType)
}

func (s *CompressSuite) TestNewCompressReader_GzipEmptyBody() {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(nil))
	req.Header.Set(contentTypeHeader, "application/json")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(reader.Close())
	}()

	data, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Empty(data)
	s.Equal("application/json", reader.DecodeType)
}
