package shared

import (
	"errors"
	"net/url"
	"strings"
)

var (
	// ErrEmptyURL is returned when the input string is empty or whitespace only.
	ErrEmptyURL = errors.New("empty url")
	// ErrEmptyHost is returned when a URL has no host after parsing.
	ErrEmptyHost = errors.New("empty host")
	// ErrNotValidScheme is returned when isSchemeRequired is true and the scheme is not http or https.
	ErrNotValidScheme = errors.New("not valid scheme")
)

// GetURL parses s into a URL. When isSchemeRequired is false, a missing scheme defaults to http.
// Host-less inputs without "://" are parsed as network-path references (//host/path).
func GetURL(s string, isSchemeRequired bool) (*url.URL, error) {

	s = strings.TrimSpace(s)
	u := &url.URL{}

	if s == "" {
		return u, ErrEmptyURL
	}

	// Parse as-is first
	parsed, err := url.Parse(s)
	if err != nil || parsed.Host == "" {
		// If there's no scheme indicator, try network-path reference to capture host
		if !strings.Contains(s, "://") {
			if p2, err2 := url.Parse("//" + s); err2 == nil {
				parsed = p2
			} else {
				// Fall back to original error if alternative parse fails
				return parsed, err2
			}
		} else if err != nil {
			return parsed, err
		}
	}

	// Validate or normalize scheme
	if isSchemeRequired {
		if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
			return parsed, ErrNotValidScheme
		}
	} else {
		// Default to http when scheme is not required and missing
		if parsed.Scheme == "" {
			parsed.Scheme = "http"
		}
	}

	if parsed.Host == "" {
		return parsed, ErrEmptyHost
	}

	return parsed, nil
}
