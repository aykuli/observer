package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage/local"
)

func Example() {
	logger := zap.NewExample()
	defer logger.Sync()
	sugar := *logger.Sugar()

	options := config.Options
	options.Restore = false
	storage, err := local.NewStorage(options, sugar)
	if err != nil {
		return
	}
	r := chi.NewRouter()
	api := APIV1{Storage: storage}

	// register endpoints
	r.Get("/", api.GetAllMetrics())
	r.Post("/update", api.UpdateFromJSON())
	r.Post("/value", api.ReadMetric())
	r.Get("/value/{metricType}/{metricName}", api.GetMetric())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", api.Update())
	r.Post("/updates", api.BatchUpdate())

	testServer := httptest.NewServer(r)
	// example gauge metrics
	valueA := 5.2
	metricID := "a_test"
	metric := models.Metric{
		ID:    metricID,
		MType: "gauge",
		Delta: nil,
		Value: &valueA,
	}

	// "/update" endpoint example
	byteData, err := json.Marshal(metric)
	if err != nil {
		fmt.Println("metric marshalling error")
		return
	}
	res, err := http.Post(testServer.URL+"/update", "application/json", bytes.NewBuffer(byteData))
	if err != nil {
		fmt.Println("metric post update error")
		return
	}
	outMetric, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("reading response body error", err.Error())
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("post /update :", string(outMetric))

	// "/value" endpoint example
	askedMetric := models.Metric{ID: metricID, MType: "gauge"}
	byteData, err = json.Marshal(askedMetric)
	if err != nil {
		fmt.Println("metric marshalling error")
		return
	}
	res, err = http.Post(testServer.URL+"/value", "application/json", bytes.NewBuffer(byteData))
	if err != nil {
		fmt.Println("metric post value error")
		return
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error decoding metric value")
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("post /value :", string(out))

	// "/update/{metricType}/{metricName}/{metricValue}" endpoint example
	res, err = http.Post(testServer.URL+"/update/gauge/b_test/1.2", "application/json", bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("error updating metric value")
		return
	}

	out, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error updating metric value")
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("post /update/gauge/b_test/1.2 :", string(out))

	// "/value/{metricType}/{metricName}" endpoint example
	res, err = http.Get(testServer.URL + "/value/gauge/" + metricID)
	if err != nil {
		fmt.Println("error getting metric value")
		return
	}
	out, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error reading metric value")
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("get value/gauge/test_metric :", string(out))

	// "/" endpoint example
	res, err = http.Get(testServer.URL)
	if err != nil {
		fmt.Println("error getting all metrics")
		return
	}
	out, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error decoding metrics")
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("get / :", string(out))

	// "/updates" endpoint example
	valueC := 1.3
	valueD := 1.4
	metrics := []models.Metric{
		{ID: "c_test", MType: "gauge", Delta: nil, Value: &valueC},
		{ID: "c_test", MType: "gauge", Delta: nil, Value: &valueD},
	}
	byteData, err = json.Marshal(metrics)

	res, err = http.Post(testServer.URL+"/updates", "application/json", bytes.NewBuffer(byteData))
	if err != nil {
		fmt.Println("error updating metrics")
		return
	}
	out, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error updating metrics")
		return
	}
	if err = res.Body.Close(); err != nil {
		fmt.Println("res body close error")
		return
	}
	fmt.Println("post /updates :", string(out))

	// Output:
	// post /update : {"id":"a_test","type":"gauge","value":5.2}
	// post /value : {"id":"a_test","type":"gauge","value":5.2}
	//
	// post /update/gauge/b_test/1.2 : {"id":"b_test","type":"gauge","value":1.2}
	//
	// get value/gauge/test_metric : 5.2
	// get / : a_test: 5.200000,
	// b_test: 1.200000
	// post /updates : [{"id":"c_test","type":"gauge","value":1.3},{"id":"c_test","type":"gauge","value":1.4}]
}
