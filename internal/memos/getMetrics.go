package memos

import (
	"math/rand/v2"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"gorono/internal/models"
)

// GetMoreMetrix получает три дополнительных метрики из gopsutil
func GetMoreMetrix() *[]models.Metrics {
	v, _ := mem.VirtualMemory() //             VirtualMemory()
	cc, _ := cpu.Counts(true)
	gaugeMap := map[string]models.Gauge{
		"TotalMemory":     models.Gauge(v.Total),
		"FreeMemory":      models.Gauge(v.Free),
		"CPUutilization1": models.Gauge(cc),
	}
	metrArray := make([]models.Metrics, 0, len(gaugeMap))
	for metrName, metrValue := range gaugeMap {
		mval := float64(metrValue)
		metr := models.Metrics{ID: metrName, MType: "gauge", Value: &mval}
		metrArray = append(metrArray, metr)
	}
	return &metrArray
}

// GetMetrixFromOS получает антайм метрики из runtime.MemStats
func GetMetrixFromOS() *[]models.Metrics {
	var mS runtime.MemStats
	runtime.ReadMemStats(&mS)
	gaugeMap := map[string]models.Gauge{
		"Alloc":         models.Gauge(mS.Alloc),
		"BuckHashSys":   models.Gauge(mS.BuckHashSys),
		"Frees":         models.Gauge(mS.Frees),
		"GCCPUFraction": models.Gauge(mS.GCCPUFraction),
		"GCSys":         models.Gauge(mS.GCSys),
		"HeapAlloc":     models.Gauge(mS.HeapAlloc),
		"HeapIdle":      models.Gauge(mS.HeapIdle),
		"HeapInuse":     models.Gauge(mS.HeapInuse),
		"HeapObjects":   models.Gauge(mS.HeapObjects),
		"HeapReleased":  models.Gauge(mS.HeapReleased),
		"HeapSys":       models.Gauge(mS.HeapSys),
		"LastGC":        models.Gauge(mS.LastGC),
		"Lookups":       models.Gauge(mS.Lookups),
		"MCacheInuse":   models.Gauge(mS.MCacheInuse),
		"MCacheSys":     models.Gauge(mS.MCacheSys),
		"MSpanInuse":    models.Gauge(mS.MSpanInuse),
		"MSpanSys":      models.Gauge(mS.MSpanSys),
		"Mallocs":       models.Gauge(mS.Mallocs),
		"NextGC":        models.Gauge(mS.NextGC),
		"NumForcedGC":   models.Gauge(mS.NumForcedGC),
		"NumGC":         models.Gauge(mS.NumGC),
		"OtherSys":      models.Gauge(mS.OtherSys),
		"PauseTotalNs":  models.Gauge(mS.PauseTotalNs),
		"StackInuse":    models.Gauge(mS.StackInuse),
		"StackSys":      models.Gauge(mS.StackSys),
		"Sys":           models.Gauge(mS.Sys),
		"TotalAlloc":    models.Gauge(mS.TotalAlloc),
		"RandomValue":   models.Gauge(rand.Float64()), // self-defined
	}

	counterMap := map[string]models.Counter{
		"PollCount": models.Counter(0), // self-defined
	}
	metrArray := make([]models.Metrics, 0, len(gaugeMap)+len(counterMap))

	for metrName, metrValue := range counterMap {
		mval := int64(metrValue)
		metr := models.Metrics{ID: metrName, MType: "counter", Delta: &mval}
		metrArray = append(metrArray, metr)
	}
	for metrName, metrValue := range gaugeMap {
		mval := float64(metrValue)
		metr := models.Metrics{ID: metrName, MType: "gauge", Value: &mval}
		metrArray = append(metrArray, metr)
	}
	return &metrArray
}

// MetrixUnMarhal - самопальный анмаршал слайса метрик, полученных от Агента
func MetrixUnMarhal(bunchOnMarsh []byte) (*[]models.Metrics, error) {
	metricArray := new([]models.Metrics) // "new" - create on heap

	metricString := string(bunchOnMarsh)

	metricString = strings.TrimPrefix(metricString, "[")
	metricString = strings.TrimSuffix(metricString, "]")
	metrics := strings.Split(metricString, "},")

	for _, metra := range metrics {
		metricFields := strings.Split(metra, ",")

		var metr = models.Metrics{}
		for _, m := range metricFields {
			m = strings.TrimPrefix(m, "{")
			m = strings.TrimSuffix(m, "}")

			field := strings.Split(m, ":")

			field[0] = strings.TrimPrefix(field[0], "\"")
			field[0] = strings.TrimSuffix(field[0], "\"")
			field[1] = strings.TrimPrefix(field[1], "\"")
			field[1] = strings.TrimSuffix(field[1], "\"")
			switch field[0] {

			case "id":
				metr.ID = field[1]
			case "type":
				metr.MType = field[1]
			case "value":
				flo, err := strconv.ParseFloat(field[1], 64)
				if err != nil {
					return nil, err
				}
				metr.Value = &flo
			case "delta":
				inta, err := strconv.ParseInt(field[1], 10, 64)
				if err != nil {
					return nil, err
				}
				metr.Delta = &inta
			}
		}
		*metricArray = append(*metricArray, metr)
	}
	return metricArray, nil
}
