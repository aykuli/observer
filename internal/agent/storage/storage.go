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

type MemStorage struct {
	gaugeMetrics   gaugeMetrics
	counterMetrics counterMetrics
	mutex          sync.RWMutex
}

func NewMemStorage() MemStorage {
	return MemStorage{
		gaugeMetrics:   map[string]float64{},
		counterMetrics: map[string]int64{},
	}
}

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

func (m *MemStorage) GetAllMetrics() []models.Metric {
	var outMetrics []models.Metric

	m.mutex.RLock()
	for k := range m.gaugeMetrics {
		v := m.gaugeMetrics[k]
		outMetrics = append(outMetrics, models.Metric{ID: k, MType: "gauge", Delta: nil, Value: &v})
	}
	for k := range m.counterMetrics {
		d := m.counterMetrics[k]
		outMetrics = append(outMetrics, models.Metric{ID: k, MType: "counter", Delta: &d, Value: nil})
	}
	m.mutex.RUnlock()

	return outMetrics
}

func (m *MemStorage) ResetCounter() {
	m.mutex.Lock()
	m.counterMetrics["PollCount"] = 0
	m.mutex.Unlock()
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
