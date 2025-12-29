package urlShortener

import (
	"errors"
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/service/hasher"
)

type UrlMap map[string]string

var urlMap = make(UrlMap)

func ShortifyUrl(url string) (string, bool, error) {
	if !isValidURL(url) {
		return "", false, errors.New("url in wrong format")
	}

	shortUrl := hasher.ShortCode8(url)

	if _, ok := urlMap[shortUrl]; ok {
		return shortUrl, false, nil
	}

	urlMap[shortUrl] = url

	return shortUrl, true, nil
}

func ResolveShortUrl(shortUrl string) (string, error) {
	url := urlMap[shortUrl]

	if url == "" {
		return "", errors.New("Url not found")
	}

	return url, nil
}

// isValidURL reports whether s is a syntactically valid absolute HTTP/HTTPS URL.
// Requirements:
// - Must parse via url.ParseRequestURI
// - Scheme must be http or https
// - Host must be non-empty
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}
