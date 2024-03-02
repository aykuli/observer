package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zapLogger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zapLogger

	return nil
}

type (
	response struct {
		statusCode int
		bodySize   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		response
	}
)

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.response.bodySize += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.response.statusCode = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		loggedRw := loggingResponseWriter{
			ResponseWriter: w,
			response:       response{statusCode: 0, bodySize: 0},
		}

		h.ServeHTTP(&loggedRw, r)

		duration := time.Since(start)

		Log.Info("Request info",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration))
		Log.Info("Response info",
			zap.Int("status_code", loggedRw.response.statusCode),
			zap.Int("body_size", loggedRw.response.bodySize))
	}

	return http.HandlerFunc(logFn)
}
