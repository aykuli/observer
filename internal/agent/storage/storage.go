// Package storage provides methods to garbage and get metrics.
package storage

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/aykuli/observer/internal/models"
)

type gaugeMetrics map[string]float64
type counterMetrics map[string]int64

// MemStorage struct keeps gauge and counter metrics
type MemStorage struct {
	gaugeMetrics   gaugeMetrics
	counterMetrics counterMetrics
	mutex          sync.RWMutex
}

// NewMemStorage returns MemStorage object.
func NewMemStorage() MemStorage {
	return MemStorage{
		gaugeMetrics:   map[string]float64{},
		counterMetrics: map[string]int64{},
	}
}

// GarbageStats gets metrics values from runtime process.
func (m *MemStorage) GarbageStats() {
	memstats := runtime.MemStats{}
	runtime.ReadMemStats(&memstats)

	m.mutex.Lock()
	m.counterMetrics["PollCount"]++

	m.gaugeMetrics["Alloc"] = float64(memstats.Alloc)
	m.gaugeMetrics["BuckHashSys"] = float64(memstats.BuckHashSys)
	m.gaugeMetrics["Frees"] = float64(memstats.Frees)
	m.gaugeMetrics["GCCPUFraction"] = memstats.GCCPUFraction
	m.gaugeMetrics["GCSys"] = float64(memstats.GCSys)
	m.gaugeMetrics["HeapAlloc"] = float64(memstats.HeapAlloc)
	m.gaugeMetrics["HeapIdle"] = float64(memstats.HeapIdle)
	m.gaugeMetrics["HeapInuse"] = float64(memstats.HeapInuse)
	m.gaugeMetrics["HeapObjects"] = float64(memstats.HeapObjects)
	m.gaugeMetrics["HeapReleased"] = float64(memstats.HeapReleased)
	m.gaugeMetrics["HeapSys"] = float64(memstats.HeapSys)
	m.gaugeMetrics["LastGC"] = float64(memstats.LastGC)
	m.gaugeMetrics["Lookups"] = float64(memstats.Lookups)
	m.gaugeMetrics["MCacheInuse"] = float64(memstats.MCacheInuse)
	m.gaugeMetrics["MCacheSys"] = float64(memstats.MCacheSys)
	m.gaugeMetrics["MSpanInuse"] = float64(memstats.MSpanInuse)
	m.gaugeMetrics["MSpanSys"] = float64(memstats.MSpanSys)
	m.gaugeMetrics["Mallocs"] = float64(memstats.Mallocs)
	m.gaugeMetrics["NextGC"] = float64(memstats.NextGC)
	m.gaugeMetrics["NumForcedGC"] = float64(memstats.NumForcedGC)
	m.gaugeMetrics["NumGC"] = float64(memstats.NumGC)
	m.gaugeMetrics["OtherSys"] = float64(memstats.OtherSys)
	m.gaugeMetrics["PauseTotalNs"] = float64(memstats.PauseTotalNs)
	m.gaugeMetrics["StackInuse"] = float64(memstats.StackInuse)
	m.gaugeMetrics["StackSys"] = float64(memstats.StackSys)
	m.gaugeMetrics["Sys"] = float64(memstats.Sys)
	m.gaugeMetrics["TotalAlloc"] = float64(memstats.TotalAlloc)
	m.gaugeMetrics["RandomValue"] = randFloat(0, 1000000)
	m.mutex.Unlock()
}

// GetSystemUtilInfo gets virtual memory metrics values.
func (m *MemStorage) GetSystemUtilInfo() {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	m.mutex.Lock()
	m.gaugeMetrics["TotalMemory"] = float64(vm.Total)
	m.gaugeMetrics["FreeMemory"] = float64(vm.Free)
	m.mutex.Unlock()

	cpuCount, err := cpu.Counts(false)
	if err != nil {
		return
	}
	vmPercent, err := cpu.Percent(0, true)
	if err != nil {
		return
	}

	m.mutex.Lock()
	for i := 0; i < cpuCount; i++ {
		mName := fmt.Sprintf("CPUutilization%d", i)
		m.gaugeMetrics[mName] = vmPercent[i]
	}
	m.mutex.Unlock()
}

// GetAllMetrics returns array of metrics.
func (m *MemStorage) GetAllMetrics() []models.Metric {
	var outMetrics = make([]models.Metric, len(m.gaugeMetrics)+len(m.counterMetrics))

	i := 0
	m.mutex.RLock()
	for k := range m.gaugeMetrics {
		v := m.gaugeMetrics[k]
		outMetrics[i] = models.Metric{ID: k, MType: "gauge", Delta: nil, Value: &v}
		i++
	}
	for k := range m.counterMetrics {
		d := m.counterMetrics[k]
		outMetrics[i] = models.Metric{ID: k, MType: "counter", Delta: &d, Value: nil}
		i++
	}
	m.mutex.RUnlock()

	return outMetrics
}

// ResetCounter sets counter PollCount to zero value as the flag
// that metrics got from previous circle of grabbing sent to storage server successfully.
func (m *MemStorage) ResetCounter() {
	m.mutex.Lock()
	m.counterMetrics["PollCount"] = 0
	m.mutex.Unlock()
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
