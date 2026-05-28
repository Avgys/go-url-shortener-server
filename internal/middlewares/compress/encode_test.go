package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (s *CompressSuite) TestGetEncodeType() {
	tests := []struct {
		name           string
		acceptEncoding string
		want           string
	}{
		{
			name:           "gzip supported",
			acceptEncoding: "gzip",
			want:           gzipType,
		},
		{
			name:           "unsupported encoding",
			acceptEncoding: "br",
			want:           noResult,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			got := getEncodeType(test.acceptEncoding)
			s.Equal(test.want, got)
		})
	}
}

func (s *CompressSuite) TestNewCompressWriter_NoEncoding() {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	writer, err := NewCompressWriter(recorder, req)
	s.Require().NoError(err)
	s.False(writer.gzipRequested)

	_, err = writer.Write([]byte("hello"))
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())
	s.Equal("hello", recorder.Body.String())
}

func (s *CompressSuite) TestNewCompressWriter_GzipSmallBodySkipped() {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(acceptEncodingHeader, gzipType)

	writer, err := NewCompressWriter(recorder, req)
	s.Require().NoError(err)
	s.True(writer.gzipRequested)

	_, err = writer.Write([]byte("hello"))
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())

	s.Equal(noResult, writer.EncodeType)
	s.Empty(recorder.Header().Get(contentEncodingHeader))
	s.Equal("hello", recorder.Body.String())
}

func (s *CompressSuite) TestNewCompressWriter_GzipLargeBody() {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(acceptEncodingHeader, gzipType)

	body := []byte(strings.Repeat("x", minGzipResponseBytes))

	writer, err := NewCompressWriter(recorder, req)
	s.Require().NoError(err)

	_, err = writer.Write(body)
	s.Require().NoError(err)
	s.Require().NoError(writer.Close())

	s.Equal(gzipType, writer.EncodeType)
	s.Equal(gzipType, recorder.Header().Get(contentEncodingHeader))

	gzr, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(gzr.Close())
	}()

	decompressed, err := io.ReadAll(gzr)
	s.Require().NoError(err)
	s.Equal(body, decompressed)
}
