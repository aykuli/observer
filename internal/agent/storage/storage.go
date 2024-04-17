package storage

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
)

type GaugeMetrics map[string]float64
type CounterMetrics map[string]int64

type MemStorage struct {
	GaugeMetrics   GaugeMetrics
	CounterMetrics CounterMetrics
	Mutex          sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GaugeMetrics:   map[string]float64{},
		CounterMetrics: map[string]int64{},
	}
}

func (m *MemStorage) GetStats() {
	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)
	m.Mutex.Lock()
	m.CounterMetrics["PollCount"]++

	m.GaugeMetrics["Alloc"] = float64(memstats.Alloc)
	m.GaugeMetrics["BuckHashSys"] = float64(memstats.BuckHashSys)
	m.GaugeMetrics["Frees"] = float64(memstats.Frees)
	m.GaugeMetrics["GCCPUFraction"] = memstats.GCCPUFraction
	m.GaugeMetrics["GCSys"] = float64(memstats.GCSys)
	m.GaugeMetrics["HeapAlloc"] = float64(memstats.HeapAlloc)
	m.GaugeMetrics["HeapIdle"] = float64(memstats.HeapIdle)
	m.GaugeMetrics["HeapInuse"] = float64(memstats.HeapInuse)
	m.GaugeMetrics["HeapObjects"] = float64(memstats.HeapObjects)
	m.GaugeMetrics["HeapReleased"] = float64(memstats.HeapReleased)
	m.GaugeMetrics["HeapSys"] = float64(memstats.HeapSys)
	m.GaugeMetrics["LastGC"] = float64(memstats.LastGC)
	m.GaugeMetrics["Lookups"] = float64(memstats.Lookups)
	m.GaugeMetrics["MCacheInuse"] = float64(memstats.MCacheInuse)
	m.GaugeMetrics["MCacheSys"] = float64(memstats.MCacheSys)
	m.GaugeMetrics["MSpanInuse"] = float64(memstats.MSpanInuse)
	m.GaugeMetrics["MSpanSys"] = float64(memstats.MSpanSys)
	m.GaugeMetrics["Mallocs"] = float64(memstats.Mallocs)
	m.GaugeMetrics["NextGC"] = float64(memstats.NextGC)
	m.GaugeMetrics["NumForcedGC"] = float64(memstats.NumForcedGC)
	m.GaugeMetrics["NumGC"] = float64(memstats.NumGC)
	m.GaugeMetrics["OtherSys"] = float64(memstats.OtherSys)
	m.GaugeMetrics["PauseTotalNs"] = float64(memstats.PauseTotalNs)
	m.GaugeMetrics["StackInuse"] = float64(memstats.StackInuse)
	m.GaugeMetrics["StackSys"] = float64(memstats.StackSys)
	m.GaugeMetrics["Sys"] = float64(memstats.Sys)
	m.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	m.GaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	m.GaugeMetrics["RandomValue"] = randFloat(0, 1000000)

	m.Mutex.Unlock()
}

func (m *MemStorage) GetSystemUtilInfo() {
	v, _ := mem.VirtualMemory()

	m.Mutex.Lock()

	m.GaugeMetrics["TotalMemory"] = float64(v.Total)
	m.GaugeMetrics["FreeMemory"] = float64(v.Free)
	m.GaugeMetrics["CPUutilization1"] = float64(v.UsedPercent)

	m.Mutex.Unlock()
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
