// Package routers provides metrics handling endpoints with chi package.
package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/compressor"
	"github.com/aykuli/observer/internal/logger"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage"
)

// MetricsRouter creates and keeps endpoints routing, middlewares them with logger, gzip functionality and handling Content-Type
func MetricsRouter(storage storage.Storage, sugarLogger *zap.Logger, options config.Config) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.WithLogging(sugarLogger))
	r.Use(compressor.GzipMiddleware)
	r.Use(middleware.AllowContentEncoding("gzip"))
	r.Use(middleware.AllowContentType("application/json", "text/html", "html/text", "text/plain"))

	v1 := handlers.APIV1{Storage: storage, Logger: sugarLogger, CryptoPrivKeyPath: options.CryptoPrivKeyPath, Key: options.Key}
	docsFs := http.FileServer(http.Dir("docs"))

	r.Route("/", func(r chi.Router) {
		//Reading endpoints
		r.Get("/", v1.GetAllMetrics())
		r.Get("/ping", v1.Ping())

		r.Route("/value", func(r chi.Router) {
			r.Post("/", v1.ReadMetric())
			r.Get("/{metricType}/{metricName}", v1.GetMetric())
		})

		//Updating endpoints
		r.Route("/update", func(r chi.Router) {
			r.Post("/", v1.UpdateFromJSON())

			r.Post("/{metricType}/{metricName}/{metricValue}", v1.Update())
		})
		r.Post("/updates/", v1.BatchUpdate())

		r.Handle("/swagger/*", http.StripPrefix("/swagger/", docsFs))
	})

	return r
}
