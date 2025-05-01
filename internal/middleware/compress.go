package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *gzipWriter) Close() {
	w.writer.Close()
}

func newGzipWriter(rsp http.ResponseWriter) (*gzipWriter, error) {
	return &gzipWriter{ResponseWriter: rsp, writer: gzip.NewWriter(rsp)}, nil
}

type gzipReader struct {
	closer io.ReadCloser
	reader *gzip.Reader
}

func (r *gzipReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *gzipReader) Close() error {
	if err := r.reader.Close(); err != nil {
		return err
	}
	return r.closer.Close()
}

func newGzipReader(rqs *http.Request) (*gzipReader, error) {
	gzr, err := gzip.NewReader(rqs.Body)
	if err != nil {
		return nil, err
	}

	return &gzipReader{rqs.Body, gzr}, nil
}

func Compress(h http.HandlerFunc) http.HandlerFunc {
	return func(rsp http.ResponseWriter, rqs *http.Request) {
		//
		// uncompress request
		//
		contentEncoding := rqs.Header.Get("Content-Encoding")
		needUncompress := strings.Contains(contentEncoding, "gzip")

		if needUncompress {
			gzr, err := newGzipReader(rqs)
			if err != nil {

				return
			}
			defer gzr.Close()

			rqs.Header.Del("Content-Encoding")
			rqs.Body = gzr
		}

		//
		// compress responce
		//
		encoding := rqs.Header.Get("Accept-Encoding")
		contentType := rqs.Header.Get("Content-Type")
		allowCompressRsp := strings.Contains(encoding, "gzip") && (contentType == "application/json" || contentType == "text/html")

		if !allowCompressRsp {
			h.ServeHTTP(rsp, rqs)
			return
		}

		gzw, err := newGzipWriter(rsp)
		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gzw.Close()

		rsp.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzw, rqs)
	}
}
