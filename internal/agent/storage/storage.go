package storage

import (
	"math/rand"
	"runtime"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics
	CounterMetrics
}

func NewMemStorage() MemStorage {
	return MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}
}

func GetStats(memStorage *MemStorage) {
	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)

	memStorage.CounterMetrics["PollCount"]++

	memStorage.GaugeMetrics["Alloc"] = float64(memstats.Alloc)
	memStorage.GaugeMetrics["BuckHashSys"] = float64(memstats.BuckHashSys)
	memStorage.GaugeMetrics["Frees"] = float64(memstats.Frees)
	memStorage.GaugeMetrics["GCCPUFraction"] = memstats.GCCPUFraction
	memStorage.GaugeMetrics["GCSys"] = float64(memstats.GCSys)
	memStorage.GaugeMetrics["HeapAlloc"] = float64(memstats.HeapAlloc)
	memStorage.GaugeMetrics["HeapIdle"] = float64(memstats.HeapIdle)
	memStorage.GaugeMetrics["HeapInuse"] = float64(memstats.HeapInuse)
	memStorage.GaugeMetrics["HeapObjects"] = float64(memstats.HeapObjects)
	memStorage.GaugeMetrics["HeapReleased"] = float64(memstats.HeapReleased)
	memStorage.GaugeMetrics["HeapSys"] = float64(memstats.HeapSys)
	memStorage.GaugeMetrics["LastGC"] = float64(memstats.LastGC)
	memStorage.GaugeMetrics["Lookups"] = float64(memstats.Lookups)
	memStorage.GaugeMetrics["MCacheInuse"] = float64(memstats.MCacheInuse)
	memStorage.GaugeMetrics["MCacheSys"] = float64(memstats.MCacheSys)
	memStorage.GaugeMetrics["MSpanInuse"] = float64(memstats.MSpanInuse)
	memStorage.GaugeMetrics["MSpanSys"] = float64(memstats.MSpanSys)
	memStorage.GaugeMetrics["Mallocs"] = float64(memstats.Mallocs)
	memStorage.GaugeMetrics["NextGC"] = float64(memstats.NextGC)
	memStorage.GaugeMetrics["NumForcedGC"] = float64(memstats.NumForcedGC)
	memStorage.GaugeMetrics["NumGC"] = float64(memstats.NumGC)
	memStorage.GaugeMetrics["OtherSys"] = float64(memstats.OtherSys)
	memStorage.GaugeMetrics["PauseTotalNs"] = float64(memstats.PauseTotalNs)
	memStorage.GaugeMetrics["StackInuse"] = float64(memstats.StackInuse)
	memStorage.GaugeMetrics["StackSys"] = float64(memstats.StackSys)
	memStorage.GaugeMetrics["Sys"] = float64(memstats.Sys)
	memStorage.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	memStorage.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	memStorage.GaugeMetrics["RandomValue"] = randFloat(0, 1000000)
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
