package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aykuli/observer/internal/storage"
)

func SendPost(client *http.Client, urlBase string, memstorage storage.MemStorage) {
	for k, v := range memstorage.GaugeMetrics {
		res, err := client.Post(getURL(urlBase, "gauge", k, v), "text/plain", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
	}

	for k, v := range memstorage.CounterMetrics {
		res, err := client.Post(getURL(urlBase, "counter", k, v), "text/plain", nil)
		if err != nil {
			log.Fatal(err)
		}

		defer res.Body.Close()

	}
}

func getURL(urlBase, mType, mName string, mValue interface{}) string {
	var mValueString string
	switch mValue.(type) {
	case float64:
		mValueString = fmt.Sprintf("%v", mValue)
	case int:
		mValueString = fmt.Sprintf("%d", mValue)
	}

	return fmt.Sprintf("%s/update/%s/%s/%v", urlBase, mType, mName, mValueString)
}
