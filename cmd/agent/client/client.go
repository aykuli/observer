// Package client provides metrics sending functionality to configured server.
package client

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"

	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/models"
	"github.com/aykuli/observer/internal/sign"

	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/compressor"
)

// Retry configuration constants if server doesn't respond
const (
	RetryCount              = 3 // retry quantity
	RetryMinWaitTimeSeconds = 1 // default wait time to sleep before retrying request.
	RetryMaxWaitTimeSeconds = 5 //  max wait time to sleep before retrying request.
)

// MetricsClient struct is used to metric sender client with settings,
// like server address string, metrics storage pointer,
// request sign key and limit of request counts to server.
type MetricsClient struct {
	ServerAddr string
	memStorage *storage.MemStorage
	signKey    string
	limit      int
}

// NewMetricsClient creates a new client for agent application.
func NewMetricsClient(config config.Config, memStorage *storage.MemStorage) *MetricsClient {
	return &MetricsClient{
		ServerAddr: "http://" + config.Address,
		memStorage: memStorage,
		signKey:    config.Key,
		limit:      config.RateLimit,
	}
}

// newRestyClient creates configured resty client for metrics client methods
func newRestyClient() *resty.Client {
	restyClient := resty.New().
		SetRetryCount(RetryCount).
		SetRetryWaitTime(RetryMinWaitTimeSeconds).
		SetRetryMaxWaitTime(RetryMaxWaitTimeSeconds).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			isConnRefused := r.StatusCode() == 0
			isServerDBErr := r.StatusCode() == http.StatusInternalServerError
			return isConnRefused || isServerDBErr
		})
	restyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Content-Encoding", "gzip")
		r.SetHeader("Accept-Encoding", "gzip")

		return nil
	})

	return restyClient
}

// SendMetrics method send metrics one by one. Request quantity might be limited.
// If not, limit will be the same as quantity of metrics
func (m *MetricsClient) SendMetrics() {
	metrics := m.memStorage.GetAllMetrics()

	var wg sync.WaitGroup
	// Создаю каналы, в них должны поместиться все метрики и все ошибки в случае, если сервер не работает и лимит равен количеству всех метрик.
	jobs := make(chan models.Metric, len(metrics))
	errs := make(chan error, len(metrics))

	// Если лимит больше количества метрик, незачем создавать лишние горутины
	if m.limit > len(metrics) {
		m.limit = len(metrics)
	}

	// Сначала создаю слушателей
	// Создаю лимитированное количество горутин, которые конкурентно слушают канал jobs
	for w := 0; w < m.limit; w++ {
		// Инкрементирую вейт группу для каждой горутины
		wg.Add(1)
		go func() {
			// Как горутина закончила работу, декремент вейт группы
			defer wg.Done()
			for metric := range jobs {
				if err := m.sendOneMetric(metric); err != nil {
					// Если ошибка случается, сохраняю в канал ошибок
					errs <- err
					return
				}
			}
		}()
	}

	// Reference https://stackoverflow.com/a/11104510
LOOP:
	// Начинаю работу над метриками
	for _, metric := range metrics {
		// Пока закидываю в канал jobs паралельно держу ухо востро, прилетит ли ошибка
		select {
		case <-errs:
			// Если прилетела ошибка, закидывать в канал метрику смысла уже нет, весь бач метрик считаю испорченным
			// и выхожу их цикла
			break LOOP
		default:
			// Ну а если все хорошо идёт, ничего не остается, как работать
			jobs <- metric
		}
	}

	// Если мы дошли сюда, значит либо все метрики обработаны, либо пришла ошибка и мы вышли из цикла `range metrics`
	close(jobs)
	// Жду, когда горутины с запросами закончат работать
	wg.Wait()
	// После окончания всех запросов канал с ошибками моно смело закрывать
	// так как уже некому паниковать, что отпавляет ошибку в закрытый канал
	close(errs)

	// Расчехляю канал ошибок
	if err := <-errs; err != nil {
		// Если была хоть одна ошибка, весь бач сичтаю испорченным и выхожу из функции
		return
	}

	// Все прошло хорошо, можно счетчик сбросить
	m.memStorage.ResetCounter()
}

func (m *MetricsClient) sendOneMetric(metric models.Metric) error {
	req := newRestyClient().R()
	req.SetHeader("Content-Type", "application/json")
	req.URL = m.ServerAddr + "/update/"
	req.Method = http.MethodPost

	marshalled, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	if m.signKey != "" {
		req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))
	}

	gzipped, err := compressor.Compress(marshalled)
	if err != nil {
		return err
	}

	if _, err = req.SetBody(gzipped).Send(); err != nil {
		return err
	}

	return nil
}

// SendBatchMetrics method send all metrics in one request.
func (m *MetricsClient) SendBatchMetrics() {
	req := newRestyClient().R()
	req.SetHeader("Content-Type", "application/json")
	req.URL = m.ServerAddr + "/updates/"
	req.Method = http.MethodPost

	metrics := m.memStorage.GetAllMetrics()
	if len(metrics) == 0 {
		return
	}

	marshalled, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Err JSON marshalling metrics with err %+v", err)
		return
	}

	if m.signKey != "" {
		req.SetHeader("HashSHA256", sign.GetHmacString(marshalled, m.signKey))
	}

	gzipped, err := compressor.Compress(marshalled)
	if err != nil {
		log.Printf("Err compressing metrics with err %+v", err)
		return
	}

	if _, err := req.SetBody(gzipped).Send(); err != nil {
		log.Printf("Err sending metrics with err %+v", err)
	} else {
		m.memStorage.ResetCounter()
	}
}
