package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handler"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/router"
	auditmocks "go-url-shortener/internal/service/audit/mocks"
	"go-url-shortener/internal/service/shortifier"
	shortifiermocks "go-url-shortener/internal/service/shortifier/mocks"
	"go-url-shortener/internal/testcommon"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

func TestShortenerSuite(t *testing.T) {
	suite.Run(t, new(ShortenerSuite))
}

type ShortenerSuite struct {
	suite.Suite
}

func (s *ShortenerSuite) getTestServer() *httptest.Server {
	log := zerolog.Nop()
	cfg, err := config.GetConfig([]string{}, &log)
	s.Require().NoError(err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)

	baseURL, err := url.Parse("http://" + ln.Addr().String())
	s.Require().NoError(err)

	cfg.RedirectDomain.Host = baseURL.Host
	cfg.RedirectDomain.Scheme = baseURL.Scheme

	store := repository.NewInMemoryStore(nil)

	ctrl := gomock.NewController(s.T())
	strGen := shortifiermocks.NewMockStringGenerator(ctrl)
	strGen.EXPECT().GetRandomString(gomock.Any()).Return("test-str").AnyTimes()

	mockAudit := auditmocks.NewMockPublisher(ctrl)
	mockAudit.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sf, err := shortifier.NewShortifier(s.T().Context(), strGen, store, mockAudit, &cfg.RedirectDomain)
	s.Require().NoError(err)

	h := handler.NewHandlers(sf, store, mockAudit)

	ts := httptest.NewUnstartedServer(router.NewRouter(h))
	ts.Listener = ln
	ts.Start()

	return ts
}

func (s *ShortenerSuite) doRequest(ts *httptest.Server, req *http.Request) *http.Response {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := client.Do(req)
	s.Require().NoError(err)

	return res
}

func (s *ShortenerSuite) storeRequest(host string, longURL string) *http.Request {
	req, err := http.NewRequest(http.MethodPost, host, strings.NewReader(longURL))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

	return req
}

func (s *ShortenerSuite) redirectRequest(host string, shortURL string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, host+"/"+shortURL, nil)
	s.Require().NoError(err)

	return req
}

func (s *ShortenerSuite) TestRouter() {
	ts := s.getTestServer()
	defer ts.Close()

	redirectURL := "http://someurl"
	host := ts.URL
	awaitedURL, err := url.JoinPath(host, "test-str")
	s.Require().NoError(err)

	testTable := []struct {
		name    string
		request *http.Request
		want    testcommon.ResponseWant
	}{
		{
			name:    "Store url",
			request: s.storeRequest(host, redirectURL),
			want:    testcommon.ResponseWant{StatusCode: http.StatusCreated, Body: awaitedURL},
		},
		{
			name:    "Redirect url",
			request: s.redirectRequest(host, "test-str"),
			want: testcommon.ResponseWant{
				StatusCode: http.StatusTemporaryRedirect,
				Body:       "",
				Headers:    map[string]string{"Location": redirectURL},
			},
		},
	}
	for _, tt := range testTable {
		s.Run(tt.name, func() {
			resp := s.doRequest(ts, tt.request)
			defer func() { _ = resp.Body.Close() }()

			testcommon.CheckResponseFields(s.T(), resp, tt.want)
		})
	}
}
