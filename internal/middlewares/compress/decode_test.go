package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func gzipBytes(t *testing.T, data string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(data)); err != nil {
		t.Fatalf("gzip write error: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close error: %v", err)
	}

	return buf.Bytes()
}

func TestGetDecodeType(t *testing.T) {
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
		t.Run(test.name, func(t *testing.T) {
			got := getDecodeType(test.contentType, test.contentEncoding)
			if got != test.want {
				t.Fatalf("getDecodeType(%q, %q) = %q, want %q", test.contentType, test.contentEncoding, got, test.want)
			}
		})
	}
}

func TestNewCompressReader_Gzip(t *testing.T) {
	body := gzipBytes(t, "hello")
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(body))
	req.Header.Set(contentTypeHeader, "application/json")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	if err != nil {
		t.Fatalf("NewCompressReader error: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close: %v", err)
		}
	}()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if got := string(decompressed); got != "hello" {
		t.Fatalf("decompressed body = %q, want %q", got, "hello")
	}

	if reader.DecodeType != gzipType {
		t.Fatalf("DecodeType = %q, want %q", reader.DecodeType, gzipType)
	}
}

func TestNewCompressReader_UnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("plain")))
	req.Header.Set(contentTypeHeader, "text/html")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	if err != nil {
		t.Fatalf("NewCompressReader error: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close: %v", err)
		}
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if got := string(data); got != "plain" {
		t.Fatalf("body = %q, want %q", got, "plain")
	}

	if reader.DecodeType != gzipType {
		t.Fatalf("DecodeType = %q, want %q", reader.DecodeType, gzipType)
	}
}

func TestNewCompressReader_GzipEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(nil))
	req.Header.Set(contentTypeHeader, "application/json")
	req.Header.Set(contentEncodingHeader, gzipType)

	reader, err := NewCompressReader(req)
	if err != nil {
		t.Fatalf("NewCompressReader error: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("reader.Close: %v", err)
		}
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if len(data) != 0 {
		t.Fatalf("expected empty body, got %q", string(data))
	}

	if reader.DecodeType != "application/json" {
		t.Fatalf("DecodeType = %q, want %q", reader.DecodeType, "application/json")
	}
}
