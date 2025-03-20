package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (rsp *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rsp.ResponseWriter.Write(b)
	rsp.responseData.size += size
	return size, err
}

func (rsp *loggingResponseWriter) WriteHeader(statusCode int) {
	rsp.ResponseWriter.WriteHeader(statusCode)
	rsp.responseData.status = statusCode
}

func Log(h http.HandlerFunc, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(rsp http.ResponseWriter, rqs *http.Request) {

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lrsp := loggingResponseWriter{
			ResponseWriter: rsp,
			responseData:   responseData,
		}

		start := time.Now()
		h.ServeHTTP(&lrsp, rqs)
		duration := time.Since(start)

		logger.Infoln(
			"uri", rqs.RequestURI,
			"method", rqs.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}
