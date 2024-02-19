package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/aykuli/observer/cmd/server/handlers"
	"github.com/aykuli/observer/internal/storage"
)

func MetricsRouter(memStorage *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	m := handlers.Metrics{MemStorage: memStorage}

	r.Route("/", func(r chi.Router) {
		r.Get("/", m.GetAllMetrics())
		r.Get("/value/{metricType}/{metricName}", m.GetMetric())

		r.Route("/update", func(r chi.Router) {
			r.Route("/{metricType}", func(r chi.Router) {
				r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
					rw.Header().Set("Content-Type", "text/plain")
					rw.WriteHeader(http.StatusNotFound)

				})
				r.Route("/{metricName}", func(r chi.Router) {
					r.Post("/", func(rw http.ResponseWriter, w *http.Request) {
						rw.Header().Set("Content-Type", "text/plain")
						rw.WriteHeader(http.StatusBadRequest)
					})

					r.Post("/{metricValue}", m.Update())
				})
			})

		})
	})

	return r
}
