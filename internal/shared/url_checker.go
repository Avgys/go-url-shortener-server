package shared

import (
	"net/url"
	"strings"

	"github.com/Avgys/go-url-shortener-server/internal/shared/common_errors"
)

func GetURL(s string, isSchemeRequired bool) (*url.URL, error) {

	s = strings.TrimSpace(s)
	u := &url.URL{}

	if s == "" {
		return u, common_errors.ErrEmptyURL
	}

	// Parse as-is first
	parsed, err := url.Parse(s)
	if err != nil || parsed.Host == "" {
		// If there's no scheme indicator, try network-path reference to capture host
		if !strings.Contains(s, "://") {
			if p2, err2 := url.Parse("//" + s); err2 == nil {
				parsed = p2
				err = nil
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
			return parsed, common_errors.ErrNotValidScheme
		}
	} else {
		// Default to http when scheme is not required and missing
		if parsed.Scheme == "" {
			parsed.Scheme = "http"
		}
	}

	if parsed.Host == "" {
		return parsed, common_errors.ErrEmptyHost
	}

	return parsed, nil
}
