package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEncodeType(t *testing.T) {
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
		t.Run(test.name, func(t *testing.T) {
			got := getEncodeType(test.acceptEncoding)
			if got != test.want {
				t.Fatalf("getEncodeType(%q) = %q, want %q", test.acceptEncoding, got, test.want)
			}
		})
	}
}

func TestNewCompressWriter_NoEncoding(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	writer, err := NewCompressWriter(recorder, req)
	if err != nil {
		t.Fatalf("NewCompressWriter error: %v", err)
	}

	if writer.EncodeWriter != nil {
		t.Fatalf("expected nil EncodeWriter when no Accept-Encoding is provided")
	}

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	if got := recorder.Body.String(); got != "hello" {
		t.Fatalf("response body = %q, want %q", got, "hello")
	}
}

func TestNewCompressWriter_Gzip(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set(acceptEncodingHeader, gzipType)

	writer, err := NewCompressWriter(recorder, req)
	if err != nil {
		t.Fatalf("NewCompressWriter error: %v", err)
	}

	if writer.EncodeWriter == nil {
		t.Fatalf("expected EncodeWriter when gzip is accepted")
	}

	if writer.EncodeType != gzipType {
		t.Fatalf("EncodeType = %q, want %q", writer.EncodeType, gzipType)
	}

	_, err = writer.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	if got := recorder.Header().Get(contentEncodingHeader); got != gzipType {
		t.Fatalf("Content-Encoding = %q, want %q", got, gzipType)
	}

	gzr, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			t.Errorf("gzr.Close: %v", err)
		}
	}()

	decompressed, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if got := string(decompressed); got != "hello" {
		t.Fatalf("decompressed body = %q, want %q", got, "hello")
	}
}
