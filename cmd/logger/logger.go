package logger

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func LogMware(h http.Handler) http.Handler {
	LogFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		respData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		var headers string
		for key, value := range r.Header {
			headers = headers + fmt.Sprintf("%s: %s ", key, value)
		}
		Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.String("headers", headers),
		)

		Log.Info(fmt.Sprintf("response for %s request", r.RequestURI),
			zap.Int("status", respData.status),
			zap.Int("size", respData.size),
		)
	}
	return http.HandlerFunc(LogFunc)
}

type (
	responseData struct {
		status int
		size   int
		body   string
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
