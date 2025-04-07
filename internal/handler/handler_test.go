package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rvkarpov/url_shortener/internal/mocks"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/testutils"

	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {
	cfg := testutils.LoadTestConfig()
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
			urlService := service.NewURLService(storage, &cfg)
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

func TestPostStringHandler(t *testing.T) {
	cfg := testutils.LoadTestConfig()
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
			urlService := service.NewURLService(storage, &cfg)
			handler := NewURLHandler(urlService, &cfg)

			router := chi.NewRouter()
			router.Post("/", handler.ProcessPostURLString)

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

func TestPostObjectHandler(t *testing.T) {
	cfg := testutils.LoadTestConfig()
	storage := mocks.NewStorageMock()

	type want struct {
		code int
		rsp  string
	}
	tests := []struct {
		name        string
		rqsData     string
		contentType string
		want        want
	}{
		{
			name:        "common",
			rqsData:     `{"url":"https://www.foo.com"}`,
			contentType: "application/json",
			want: want{
				code: 201,
				rsp:  `{"result":"http://localhost:8080/6ySFbLgd"}`,
			},
		},
		{
			name:        `no "url" key`,
			rqsData:     `{"unknown_key" : "https://www.foo.com"}`,
			contentType: "application/json",
			want: want{
				code: 400,
				rsp:  "url not specified\n",
			},
		},
		{
			name:        "invalid json",
			rqsData:     `{"https://www.foo.com"}`,
			contentType: "application/json",
			want: want{
				code: 400,
				rsp:  "invalid json\n",
			},
		},
		{
			name:        "empty obj",
			contentType: "application/json",
			rqsData:     "",
			want: want{
				code: 400,
				rsp:  "invalid json\n",
			},
		},
		{
			name:        "not json",
			rqsData:     "bar",
			contentType: "plain/text",
			want: want{
				code: 400,
				rsp:  "incorrect content type\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService := service.NewURLService(storage, &cfg)
			handler := NewURLHandler(urlService, &cfg)

			router := chi.NewRouter()
			router.Post("/api/shorten", handler.ProcessPostURLObject)

			rqs := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(test.rqsData))
			rqs.Header.Set("Content-Type", test.contentType)

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

func TestPostBatchHandler(t *testing.T) {
	cfg := testutils.LoadTestConfig()
	storage := mocks.NewStorageMock()

	type want struct {
		code int
		rsp  string
	}
	tests := []struct {
		name        string
		rqsData     string
		contentType string
		want        want
	}{
		{
			name: "common",
			rqsData: `[{"correlation_id" : "id1", "original_url" : "https://www.foo1.com"},
			           {"correlation_id" : "id2", "original_url" : "https://www.foo2.com"}]`,
			contentType: "application/json",
			want: want{
				code: 201,
				rsp: strings.Join(strings.Fields(
					`[{"correlation_id" : "id1", "short_url" : "http://localhost:8080/XTuTMq3X"},
				      {"correlation_id" : "id2", "short_url" : "http://localhost:8080/FlT-CpRc"}]`), ""),
			},
		},
		{
			name:        "not arr",
			rqsData:     `{"id1":"https://www.foo1.com"}`,
			contentType: "application/json",
			want: want{
				code: 400,
				rsp:  "invalid json\n",
			},
		},
		{
			name:        "empty arr",
			contentType: "application/json",
			rqsData:     "[]",
			want: want{
				code: 400,
				rsp:  "empty batch\n",
			},
		},
		{
			name:        "not json",
			rqsData:     "bar",
			contentType: "plain/text",
			want: want{
				code: 400,
				rsp:  "incorrect content type\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService := service.NewURLService(storage, &cfg)
			handler := NewURLHandler(urlService, &cfg)

			router := chi.NewRouter()
			router.Post("/api/shorten/batch", handler.ProcessPostURLBatch)

			rqs := httptest.NewRequest(
				http.MethodPost,
				"/api/shorten/batch",
				bytes.NewBufferString(test.rqsData),
			)
			rqs.Header.Set("Content-Type", test.contentType)

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
