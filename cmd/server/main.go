package main

import (
	"log"
	"net/http"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/storage"
)

// Доработайте сервер так, чтобы в ответ на запрос
// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
// он возвращал текущее значение метрики в текстовом виде со статусом http.StatusOK.
// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.
// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен о
// тдавать HTML-страницу со списком имён и
// значений всех известных ему на текущий момент метрик.
func main() {
	memStorage := storage.MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}

	log.Fatal(http.ListenAndServe(`localhost:8080`, routers.MetricsRouter(&memStorage)))
}
