package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rvkarpov/url_shortener/internal/config"
	"github.com/rvkarpov/url_shortener/internal/mocks"
	"github.com/rvkarpov/url_shortener/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {
	cfg := config.LoadConfig()
	storage := mocks.NewStorageMock()
	storage.AddTestData("oeapEa", "https://www.foo.com")

	type want struct {
		code     int
		location string
		rsp      string
	}
	tests := []struct {
		name string
		rqs  string
		want want
	}{
		{
			name: "existed short URL",
			rqs:  "/oeapEa",
			want: want{
				code:     307,
				location: "https://www.foo.com",
				rsp:      "",
			},
		},
		{
			name: "not existed short URL",
			rqs:  "/3Zgnmj",
			want: want{
				code:     400,
				location: "",
				rsp:      "URL not found\n",
			},
		},
		{
			name: "invalid short URL",
			rqs:  "/incorrect-short-url",
			want: want{
				code:     400,
				location: "",
				rsp:      "URL not found\n",
			},
		},
		{
			name: "empty path",
			rqs:  "/",
			want: want{
				code:     404,
				location: "",
				rsp:      "404 page not found\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService := service.NewURLService(storage)
			handler := NewURLHandler(urlService, &cfg)

			router := chi.NewRouter()
			router.Get("/{URL}", handler.ProcessGet)

			rqs := httptest.NewRequest(http.MethodGet, test.rqs, nil)
			rsp := httptest.NewRecorder()
			router.ServeHTTP(rsp, rqs)

			res := rsp.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.location, res.Header.Get("Location"))
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, test.want.rsp, string(resBody))
		})
	}
}

func TestPostHandler(t *testing.T) {
	cfg := config.LoadConfig()
	storage := mocks.NewStorageMock()

	type want struct {
		code int
		rsp  string
	}
	tests := []struct {
		name    string
		rqsData string
		want    want
	}{
		{
			name:    "common",
			rqsData: "https://www.foo.com",
			want: want{
				code: 201,
				rsp:  "http://localhost:8080/6ySFbLgd",
			},
		},
		{
			name:    "incorrect URL",
			rqsData: "www.foo.com",
			want: want{
				code: 400,
				rsp:  "invalid URL\n",
			},
		},
		{
			name:    "empty URL",
			rqsData: "",
			want: want{
				code: 400,
				rsp:  "empty URL\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService := service.NewURLService(storage)
			handler := NewURLHandler(urlService, &cfg)

			router := chi.NewRouter()
			router.Post("/", handler.ProcessPost)

			rqs := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.rqsData))
			rsp := httptest.NewRecorder()
			router.ServeHTTP(rsp, rqs)

			res := rsp.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, test.want.rsp, string(resBody))
		})
	}
}
