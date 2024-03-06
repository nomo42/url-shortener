package gz_encode

import (
	"compress/gzip"
	"fmt"
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

func (w gzipWriter) Write(b []byte) (int, error) {
	cType := w.Header().Get("Content-Type")
	if (cType == "application/json" || cType == "text/html") && w.Writer != nil {
		w.f()
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func GzipMware(h http.Handler) http.Handler {
	EncodeFunc := func(w http.ResponseWriter, r *http.Request) {

		gzwr := gzipWriter{
			ResponseWriter: w,
		}
		if strings.Contains(r.Header.Values("Accept-Encoding")[0], "gzip") {
			gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
			if err != nil {
				_, _ = io.WriteString(w, err.Error())
			}
			gzwr.Writer = gz
			gzwr.f = sync.OnceFunc(func() {
				fmt.Println("HEY2")
				w.Header().Set("Content-Encoding", "gzip")
			})
			fmt.Println("HEY")
			defer func() {
				cType := w.Header().Get("Content-Type")
				if cType == "application/json" || cType == "text/html" {
					gz.Close()
				}
			}()
		}
		fmt.Println(r.Header.Values("Accept-Encoding"))
		h.ServeHTTP(gzwr, r)

	}
	return http.HandlerFunc(EncodeFunc)
}
