package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aykuli/observer/internal/storage"
)

func checkType(metricType string) bool {
	if metricType == "gauge" || metricType == "counter" {
		return true
	}

	return false
}

func UpdateRuntime(memstorage storage.MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "Only POST method allowed", http.StatusMethodNotAllowed)
			return
		}

		var metricType, metricName, metricValue string
		for i, value := range strings.Split(r.URL.Path, "/") {
			switch i {
			case 2:
				metricType = value
			case 3:
				metricName = value
			case 4:
				metricValue = value
			}
		}

		if !checkType(metricType) {
			http.Error(rw, "Metric type is wrong", http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(rw, "Metric type is wrong", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, "Metric value is wrong", http.StatusBadRequest)
				return

			}
			memstorage.GaugeMetrics[metricName] = value
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, "Metric value is wrong", http.StatusBadRequest)
				return

			}

			memstorage.CounterMetrics[metricName] = +value
		default:
			http.Error(rw, "No such metric type", http.StatusNotFound)
			return

		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}
