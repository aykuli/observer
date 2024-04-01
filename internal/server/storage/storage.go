package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/aykuli/observer/internal/server/config"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics   GaugeMetrics   `json:"gauge_metrics"`
	CounterMetrics CounterMetrics `json:"counter_metrics"`
}

func (m *MemStorage) Load() error {
	// 1.Есть ли файл с конфигурацией сервера, где хранится путь к файлу с метриками, заданный в предыдущем запуске пользователем
	// Если есть - открываем, если нет - создаем и открываем.
	modeVal, err := strconv.ParseUint("0777", 8, 32)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(config.Options.Path, os.O_RDWR|os.O_CREATE, os.FileMode(modeVal))
	if err != nil {
		return err
	}

	defer file.Close()
	defer m.saveUserPath(file)

	reader := bufio.NewScanner(file)
	reader.Scan()
	metricsFilePath := reader.Bytes()

	if len(metricsFilePath) == 0 {
		return nil
	}

	err = m.loadMetricsFromFile(string(metricsFilePath))
	if err != nil {
		return err
	}

	return nil
}

func (m *MemStorage) saveUserPath(file *os.File) error {
	err := file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(config.Options.FileStoragePath))
	if err != nil {
		return err
	}

	return nil
}

func (m *MemStorage) loadMetricsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil // if file corrupted or didnt exists - anyway it will be removed
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	return nil
}

func (m *MemStorage) SaveMetricsPeriodically() {
	collectTicker := time.NewTicker(time.Duration(config.Options.StoreInterval) * time.Second)
	collectQuit := make(chan struct{})
	for {
		select {
		case <-collectTicker.C:
			m.SaveMetricsToFile()
		case <-collectQuit:
			collectTicker.Stop()
		}
	}
}

func (m *MemStorage) SaveMetricsToFile() error {
	modeVal, err := strconv.ParseUint("0666", 8, 32)
	if err != nil {
		return nil
	}

	file, err := os.OpenFile(config.Options.FileStoragePath, os.O_WRONLY|os.O_CREATE, os.FileMode(modeVal))
	if err != nil {
		return err
	}

	if err = file.Truncate(0); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if _, err = file.Write(data); err != nil {
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	return nil
}
