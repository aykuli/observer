// Package compressor provides middleware for zipping data.
package compressor

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aykuli/observer/internal/models"
)

func Example() {
	r := chi.NewRouter()
	r.Use(GzipMiddleware)
	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		var gotMetric models.Metric
		if err := json.NewDecoder(r.Body).Decode(&gotMetric); err != nil {
			fmt.Println("cannot decode json request body")
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		fmt.Printf("got metric: { id: %s, type: %s, value: %f }", gotMetric.ID, gotMetric.MType, *gotMetric.Value)

		w.WriteHeader(http.StatusOK)
		r.Body.Close()
	})

	testServer := httptest.NewServer(r)
	value := 1.2
	someMetric := models.Metric{ID: "test", MType: "gauge", Value: &value}
	byteData, err := json.Marshal(someMetric)
	if err != nil {
		fmt.Println("metric marshalling error")
		return
	}

	req := resty.New().R()
	req.URL = testServer.URL + "/test"
	req.Method = http.MethodPost
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("Content-Encoding", "gzip")
	gzipped, err := Compress(byteData)
	if err != nil {
		fmt.Println("gzip compressing error")
		return
	}
	req.SetBody(gzipped)
	_, err = req.Send()
	if err != nil {
		fmt.Println("req sending error")
		return
	}
	time.Sleep(time.Microsecond)

	// Output:
	// got metric: { id: test, type: gauge, value: 1.200000 }
}

func TestCompress(t *testing.T) {
	str := []byte(`Lorem ipsum dolor sit amet, 
consectetur adipiscing elit, sed do eiusmod tempor incididunt 
ut labore et dolore magna aliqua. Ut enim ad minim veniam, 
quis nostrud exercitation ullamco laboris nisi ut aliquip 
ex ea commodo consequat. Duis aute irure dolor in reprehenderit 
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 
Excepteur sint occaecat cupidatat non proident, sunt in culpa 
qui officia deserunt mollit anim id est laborum.`)
	res, err := Compress(str)
	require.NoError(t, err)
	assert.Less(t, binary.Size(res), binary.Size(str))
}
