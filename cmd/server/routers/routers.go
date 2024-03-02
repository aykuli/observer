package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/compressor"
	"github.com/aykuli/observer/internal/server/storage"
)

func MetricsRouter(storage storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(compressor.GzipMiddleware)
	r.Use(middleware.AllowContentEncoding("gzip"))
	r.Use(middleware.AllowContentType("application/json", "text/html", "html/text", "text/plain"))

	v1 := handlers.APIV1{Storage: storage}

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
	})

	return r
}
