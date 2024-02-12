package client

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/storage"
)

type MerticsClient struct {
	ServerAddr string
	MemStorage storage.MemStorage
}

func (m *MerticsClient) SendMetrics(req *resty.Request) {
	for k, v := range m.MemStorage.GaugeMetrics {
		err := post(req, Options{m.ServerAddr, "gauge", k, fmt.Sprintf("%v", v)})
		if err != nil {
			log.Fatal(err)
		}
	}

	for k, v := range m.MemStorage.CounterMetrics {
		err := post(req, Options{m.ServerAddr, "counter", k, fmt.Sprintf("%v", v)})
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Options struct {
	serverAddr, mType, mName, mValue string
}

func post(req *resty.Request, options Options) error {
	_, err := req.SetPathParams(map[string]string{
		"mType": options.mType,
		"mName": options.mName,
		"mValue": fmt.Sprintf("%v", options.
			mValue),
	}).Post(options.serverAddr + "/update/{mType}/{mName}/{mValue}")

	return err
}
