package gzencode

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

type gzipReadCloser struct {
	io.ReadCloser
	gzReader io.Reader
}

func (w gzipWriter) Write(b []byte) (int, error) {
	cType := w.Header().Get("Content-Type")
	if (cType == "application/json" || cType == "text/html") && w.Writer != nil {
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (rc gzipReadCloser) Read(p []byte) (n int, err error) {
	n, err = rc.gzReader.Read(p)
	return n, err
}

func (rc gzipReadCloser) Close() error {
	err := rc.ReadCloser.Close()
	return err
}

func GzipWriteMware(h http.Handler) http.Handler {
	EncodeFunc := func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		for _, v := range r.Header.Values("Accept-Encoding") {
			if strings.Contains(v, "gzip") {
				ok = true
				break
			}
		}

		if ok {
			gzwr := gzipWriter{
				ResponseWriter: w,
			}
			gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
			if err != nil {
				_, _ = io.WriteString(w, err.Error())
				w.WriteHeader(http.StatusInternalServerError)

			}
			gzwr.Writer = gz
			defer func() {
				cType := w.Header().Get("Content-Type")
				if cType == "application/json" || cType == "text/html" {
					gz.Close()
				}

			}()
			h.ServeHTTP(gzwr, r)
			return
		}
		h.ServeHTTP(w, r)

	}
	return http.HandlerFunc(EncodeFunc)
}
