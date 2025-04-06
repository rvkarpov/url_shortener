package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/rvkarpov/url_shortener/internal/handler"
	"github.com/rvkarpov/url_shortener/internal/mocks"
	"github.com/rvkarpov/url_shortener/internal/service"
	"github.com/rvkarpov/url_shortener/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func compress(in string) string {
	buf := bytes.NewBuffer(nil)
	gzw := gzip.NewWriter(buf)
	gzw.Write([]byte(in))
	gzw.Close()

	return buf.String()
}

func TestPostObjectHandler(t *testing.T) {
	cfg := testutils.LoadTestConfig()
	storage := mocks.NewStorageMock()

	type want struct {
		code int
		rsp  string
	}
	tests := []struct {
		name            string
		rqsData         string
		contentType     string
		contentEncoding string
		acceptEncoding  string
		want            want
	}{
		{
			name:            "common",
			rqsData:         compress(`{"url":"https://www.foo.com"}`),
			contentType:     "application/json",
			contentEncoding: "gzip",
			acceptEncoding:  "gzip",
			want: want{
				code: 201,
				rsp:  compress(`{"result":"http://localhost:8080/6ySFbLgd"}`),
			},
		},
		{
			name:            "no compress",
			rqsData:         `{"url":"https://www.foo.com"}`,
			contentType:     "application/json",
			contentEncoding: "",
			acceptEncoding:  "",
			want: want{
				code: 201,
				rsp:  `{"result":"http://localhost:8080/6ySFbLgd"}`,
			},
		},
		{
			name:            "compress_rsp",
			rqsData:         `{"url":"https://www.foo.com"}`,
			contentType:     "application/json",
			contentEncoding: "",
			acceptEncoding:  "gzip",
			want: want{
				code: 201,
				rsp:  compress(`{"result":"http://localhost:8080/6ySFbLgd"}`),
			},
		},
		{
			name:            "compressed_rqs",
			rqsData:         compress(`{"url":"https://www.foo.com"}`),
			contentType:     "application/json",
			contentEncoding: "gzip",
			acceptEncoding:  "",
			want: want{
				code: 201,
				rsp:  `{"result":"http://localhost:8080/6ySFbLgd"}`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlService := service.NewURLService(storage)

			handler := handler.NewURLHandler(urlService, &cfg)
			router := chi.NewRouter()
			router.Post("/api/shorten", Compress(handler.ProcessPostURLObject))

			rqs := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(test.rqsData))
			rqs.Header.Set("Content-Type", test.contentType)
			rqs.Header.Set("Content-Encoding", test.contentEncoding)
			rqs.Header.Set("Accept-Encoding", test.acceptEncoding)

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
