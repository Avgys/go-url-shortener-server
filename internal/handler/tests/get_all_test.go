package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"

	"go-url-shortener/internal/model/responses"
)

func (s *HandlerSuite) Test_handlers_GetAll() {
	urls := []string{
		"http://long-url-1.com",
		"http://long-url-2.com",
	}

	r := s.getRouter(&innerStructure{
		strGen: &mockStrGenSequence{values: []string{"short-1", "short-2"}},
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	s.Require().NoError(err)

	client := &http.Client{Jar: jar}

	getAllURL, err := url.JoinPath(ts.URL, "api", "user", "urls")
	s.Require().NoError(err)

	for _, originalURL := range urls {
		req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(originalURL))
		s.Require().NoError(err)
		req.Header.Set("Content-Type", "text/plain")

		res, err := client.Do(req)
		s.Require().NoError(err)
		res.Body.Close()

		s.Equal(http.StatusCreated, res.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, getAllURL, nil)
	s.Require().NoError(err)

	res, err := client.Do(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	s.Equal(http.StatusOK, res.StatusCode)

	var resp []responses.URLPair
	err = json.NewDecoder(res.Body).Decode(&resp)
	s.Require().NoError(err)
	s.Len(resp, len(urls))

	originals := map[string]bool{}
	for _, u := range urls {
		originals[u] = true
	}

	for _, item := range resp {
		s.True(originals[item.OriginalURL])
		parsed, err := url.Parse(item.ShortURL)
		s.Require().NoError(err)
		s.NotEmpty(strings.TrimPrefix(parsed.Path, "/"))
	}
}
