package routers

import (
	"compress/gzip"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/server/storage"
)

func MetricsRouter(memStorage *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentEncoding("gzip"))
	r.Use(middleware.SetHeader("Accept-Encoding", "gzip"))
	r.Use(middleware.Compress(gzip.BestCompression, "application/json", "text/html", "html/text", "text/plain"))
	r.Use(middleware.AllowContentType("application/json", "text/html", "html/text", "text/plain"))

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

			r.Route("/{metricType}", func(r chi.Router) {
				r.Post("/", m.NotFound())

				r.Route("/{metricName}", func(r chi.Router) {
					r.Post("/", m.BadRequest())
					r.Post("/{metricValue}", m.Update())
				})
			})

		})
	})

	return r
}
