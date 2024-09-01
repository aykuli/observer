// Package logger provides logging string on Stdout to requests.
package logger

import (
	"net/http"

	"go.uber.org/zap"
)

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

func WithLogging(logger zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			loggedRw := loggingResponseWriter{
				ResponseWriter: w,
				response:       &response{statusCode: 0, bodySize: 0},
			}

			h.ServeHTTP(&loggedRw, r)

			logger.Infoln(r.Method, loggedRw.response.statusCode, r.RequestURI, "size: ", loggedRw.response.bodySize)
		}
		return http.HandlerFunc(logFn)
	}
}
