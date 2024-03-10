package gzencode

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
	f      func()
}

type gzipReadCloser struct {
	io.ReadCloser
	gzReader io.Reader
}

func (w gzipWriter) Write(b []byte) (int, error) {
	cType := w.Header().Get("Content-Type")
	if (cType == "application/json" || cType == "text/html") && w.Writer != nil {
		w.f()
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

//func GzipReadMware(h http.Handler) http.Handler {
//	DecodeFunc := func(w http.ResponseWriter, r *http.Request) {
//		values := r.Header.Values("Content-encoding")
//		var ok bool
//		for _, v := range values {
//			if strings.Contains(v, "gzip") {
//				ok = true
//				break
//			}
//		}
//
//		if ok {
//			var gzBody gzipReadCloser
//			gzReader, err := gzip.NewReader(r.Body)
//			if err != nil && err != io.EOF {
//				w.WriteHeader(http.StatusInternalServerError)
//			}
//			gzBody.gzReader = gzReader
//			gzBody.ReadCloser = r.Body
//			r.Body = gzBody
//		}
//
//		h.ServeHTTP(w, r)
//	}
//
//	return http.HandlerFunc(DecodeFunc)
//}

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
			gzwr.f = sync.OnceFunc(func() {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Encoding", "gzip")
			})
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
