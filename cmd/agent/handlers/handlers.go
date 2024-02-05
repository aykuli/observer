package handlers

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/storage"
)

func SendPost(req *resty.Request, urlBase string, memStorage storage.MemStorage) {
	for k, v := range memStorage.GaugeMetrics {
		err := post(req, Options{urlBase, "gauge", k, fmt.Sprintf("%v", v)})
		if err != nil {
			log.Fatal(err)
		}
	}

	for k, v := range memStorage.CounterMetrics {
		err := post(req, Options{urlBase, "counter", k, fmt.Sprintf("%v", v)})
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Options struct {
	urlBase, mType, mName, mValue string
}

func post(req *resty.Request, options Options) error {
	_, err := req.SetPathParams(map[string]string{
		"mType": options.mType,
		"mName": options.mName,
		"mValue": fmt.Sprintf("%v", options.
			mValue),
	}).Post(options.urlBase + "/update/{mType}/{mName}/{mValue}")

	return err
}
