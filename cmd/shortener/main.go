package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/service/urlShortener"
)

func main() {

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", webhook)
	return http.ListenAndServe(`:8080`, http.HandlerFunc(webhook))
}

func webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ShortifyUrl(w, r)
	case http.MethodGet:
		Redirect(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	if url, err = GetUriParam(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if url, err = urlShortener.ResolveShortUrl(url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)

	WriteResponse(w, "", http.StatusTemporaryRedirect)
}

func ShortifyUrl(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error

	if url, err = GetRequestBody(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isCreated := false

	if url, isCreated, err = urlShortener.ShortifyUrl(url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var status int
	if isCreated {
		status = 201
	} else {
		status = http.StatusOK
	}

	WriteResponse(w, url, status)
}

func GetUriParam(r *http.Request) (string, error) {

	segments := strings.Split(r.URL.Path, "/")
	param := strings.Trim(segments[1], "/")

	if param == "" {
		return "", fmt.Errorf("Uri param %v not found", param)
	}

	return param, nil
}

func GetRequestBody(r *http.Request) (string, error) {

	if r.Header.Get("Content-type") != "text/plain" {
		return "", errors.New("Wrong content-type")
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

func WriteResponse(w http.ResponseWriter, text string, code int) {

	w.WriteHeader(code)
	if text != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(text))
	}
}
