package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/compressor"
	"github.com/aykuli/observer/internal/server/storage"
)

func MetricsRouter(memStorage *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(
		middleware.Logger,
		compressor.GzipMiddleware,
		middleware.AllowContentEncoding("gzip"),
		middleware.AllowContentType("application/json", "text/html", "html/text", "text/plain"),
	)

	m := handlers.Metric{MemStorage: memStorage}

	r.Route("/", func(r chi.Router) {
		//Reading endpoints
		r.Get("/", m.GetAllMetrics())
		r.Get("/ping", m.Ping())

		r.Route("/value", func(r chi.Router) {
			r.Post("/", m.ReadMetric())
			r.Get("/{metricType}/{metricName}", m.GetMetric())
		})

		//Updating endpoints
		r.Route("/update", func(r chi.Router) {
			r.Post("/", m.UpdateFromJSON())

			r.Post("/{metricType}/{metricName}/{metricValue}", m.Update())
		})

		r.Post("/updates/", m.BatchUpdate())
	})

	return r
}
