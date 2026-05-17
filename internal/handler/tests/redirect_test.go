package handler_test

import (
	"net/http"
	"net/http/httptest"

	dbmodel "go-url-shortener/internal/model/db"
	"go-url-shortener/internal/repository"
	"go-url-shortener/internal/testcommon"
)

func (s *HandlerSuite) Test_handlers_Redirect() {
	tests := []struct {
		name             string
		url              string
		defaultStructure *innerStructure
		want             testcommon.ResponseWant
	}{
		{
			name: "Get redirect",
			url:  "/short-url",
			defaultStructure: &innerStructure{
				store: repository.NewInMemoryStore([]*dbmodel.DBURL{{OriginalURL: "full-url", ShortURL: "short-url"}}),
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusTemporaryRedirect,
				Headers:    map[string]string{"Location": "full-url"},
			},
		},
		{
			name: "Not found url",
			url:  "/" + testcommon.ShortHash,
			defaultStructure: &innerStructure{
				strGen: &testcommon.MockStrGen{},
			},
			want: testcommon.ResponseWant{
				StatusCode: http.StatusNotFound,
			},
		},
		{
			name: "No url param",
			url:  "",
			want: testcommon.ResponseWant{
				StatusCode: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, testHost.String()+tt.url, nil)

			recorder := httptest.NewRecorder()
			r := s.getRouter(tt.defaultStructure)

			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer func() { _ = res.Body.Close() }()

			testcommon.CheckResponseFields(s.T(), res, tt.want)
		})
	}
}
