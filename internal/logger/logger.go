// Package logger provides logging string on Stdout to requests.
package logger

import (
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates  zap.Logger configured to show logs in stdout
func New() (*zap.Logger, error) {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	return cfg.Build()
}

type response struct {
	statusCode int
	bodySize   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	response *response
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.response.bodySize += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.response.statusCode = statusCode
}

func WithLogging(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			loggedRw := loggingResponseWriter{
				ResponseWriter: w,
				response:       &response{statusCode: 0, bodySize: 0},
			}

			h.ServeHTTP(&loggedRw, r)

			logger.Info(
				"server",
				zap.String("Method", r.Method),
				zap.Int("Status code", loggedRw.response.statusCode),
				zap.String("URI", r.RequestURI),
				zap.Int("Size", loggedRw.response.bodySize))
		}
		return http.HandlerFunc(logFn)
	}
}
