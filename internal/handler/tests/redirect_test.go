package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Avgys/go-url-shortener-server/internal/config"
	"github.com/Avgys/go-url-shortener-server/internal/model"
	"github.com/Avgys/go-url-shortener-server/internal/repository"
	"github.com/Avgys/go-url-shortener-server/internal/testcommon"
)

func Test_handlers_Redirect(t *testing.T) {

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
				store:  repository.NewInMemoryStore([]*model.DBURL{{OriginalURL: "full-url", ShortURL: "short-url"}}),
				config: &config.Config{AppURL: testHost},
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
		t.Run(tt.name, func(t *testing.T) {

			//Init
			req := httptest.NewRequest(http.MethodGet, testHost.String()+tt.url, nil)

			recorder := httptest.NewRecorder()
			r := getRouter(tt.defaultStructure)

			//Run
			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			//Check
			testcommon.CheckResponseFields(t, res, tt.want)
		})
	}
}
