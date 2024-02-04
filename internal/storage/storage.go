package storage

import "runtime"

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics
	CounterMetrics
}

func GetStats(memstorage *MemStorage) {
	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)

	memstorage.CounterMetrics["PollCount"]++

	memstorage.GaugeMetrics["Alloc"] = float64(memstats.Alloc)
	memstorage.GaugeMetrics["BuckHashSys"] = float64(memstats.BuckHashSys)
	memstorage.GaugeMetrics["Frees"] = float64(memstats.Frees)
	memstorage.GaugeMetrics["GCCPUFraction"] = memstats.GCCPUFraction
	memstorage.GaugeMetrics["GCSys"] = float64(memstats.GCSys)
	memstorage.GaugeMetrics["HeapAlloc"] = float64(memstats.HeapAlloc)
	memstorage.GaugeMetrics["HeapIdle"] = float64(memstats.HeapIdle)
	memstorage.GaugeMetrics["HeapInuse"] = float64(memstats.HeapInuse)
	memstorage.GaugeMetrics["HeapObjects"] = float64(memstats.HeapObjects)
	memstorage.GaugeMetrics["HeapReleased"] = float64(memstats.HeapReleased)
	memstorage.GaugeMetrics["HeapSys"] = float64(memstats.HeapSys)
	memstorage.GaugeMetrics["LastGC"] = float64(memstats.LastGC)
	memstorage.GaugeMetrics["Lookups"] = float64(memstats.Lookups)
	memstorage.GaugeMetrics["MCacheInuse"] = float64(memstats.MCacheInuse)
	memstorage.GaugeMetrics["MCacheSys"] = float64(memstats.MCacheSys)
	memstorage.GaugeMetrics["MSpanInuse"] = float64(memstats.MSpanInuse)
	memstorage.GaugeMetrics["MSpanSys"] = float64(memstats.MSpanSys)
	memstorage.GaugeMetrics["Mallocs"] = float64(memstats.Mallocs)
	memstorage.GaugeMetrics["NextGC"] = float64(memstats.NextGC)
	memstorage.GaugeMetrics["NumForcedGC"] = float64(memstats.NumForcedGC)
	memstorage.GaugeMetrics["NumGC"] = float64(memstats.NumGC)
	memstorage.GaugeMetrics["OtherSys"] = float64(memstats.OtherSys)
	memstorage.GaugeMetrics["PauseTotalNs"] = float64(memstats.PauseTotalNs)
	memstorage.GaugeMetrics["StackInuse"] = float64(memstats.StackInuse)
	memstorage.GaugeMetrics["StackSys"] = float64(memstats.StackSys)
	memstorage.GaugeMetrics["Sys"] = float64(memstats.Sys)
	memstorage.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	memstorage.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	memstorage.GaugeMetrics["RandomValue"] = 0
}
