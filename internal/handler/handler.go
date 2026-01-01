package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/handler/config"
)

type Shortifier interface {
	ResolveShortUrl(shortUrl string) (string, error)
	ShortifyUrl(url string) (string, bool, error)
}

type handlers struct {
	shortifier Shortifier
}

func Serve(cfg config.Config, shortifier Shortifier) error {

	h := &handlers{shortifier: shortifier}

	router := newRouter(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	return srv.ListenAndServe()
}

func newRouter(h *handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.ShortifyUrl)
	mux.HandleFunc("GET /{url}", h.Redirect)

	return mux
}

func (h *handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	if url, err = getUriParam(r); err != nil {
		writeError(w, r, err)
		return
	}

	if url, err = h.shortifier.ResolveShortUrl(url); err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", url)

	writeResponse(w, "", http.StatusTemporaryRedirect)
}

func (h *handlers) ShortifyUrl(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	if url, err = getRequestBody(r); err != nil {
		writeError(w, r, err)
		return
	}

	isCreated := false

	if url, isCreated, err = h.shortifier.ShortifyUrl(url); err != nil {
		writeError(w, r, err)
		return
	}

	var status int
	if isCreated {
		status = 201
	} else {
		status = http.StatusOK
	}

	writeResponse(w, url, status)
}

func getUriParam(r *http.Request) (string, error) {
	urlParam := r.PathValue("url")

	if urlParam == "" {
		return "", errors.ErrUnsupported
	}

	return urlParam, nil
}

func getRequestBody(r *http.Request) (string, error) {

	if !strings.Contains(r.Header.Get("Content-type"), "text/plain") {
		return "", fmt.Errorf("Wrong content-type %v", r.Header.Get("Content-type"))
	}

	buffer := make([]byte, 128)
	readBytes := 0
	var err error

	if readBytes, err = r.Body.Read(buffer); err != nil && err != io.EOF {
		fmt.Printf("Got error %v\n", err)

		return "", errors.New("Got error reading url")
	}

	url := string(buffer[:readBytes])

	return url, nil
}

func writeResponse(w http.ResponseWriter, text string, code int) {

	w.WriteHeader(code)
	if text != "" {
		w.Write([]byte(text))
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	payload := struct {
		Method      string      `json:"method"`
		Path        string      `json:"path"`
		Query       string      `json:"query"`
		Headers     http.Header `json:"headers"`
		ContentType string      `json:"contentType"`
		RemoteAddr  string      `json:"remoteAddr"`
		Error       string      `json:"error"`
	}{
		Method:      r.Method,
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
		Headers:     r.Header,
		ContentType: r.Header.Get("Content-Type"),
		RemoteAddr:  r.RemoteAddr,
		Error:       err.Error(),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(payload)

	http.Error(w, err.Error(), http.StatusBadRequest)
}
