package memos

import (
	"math/rand/v2"
	"runtime"
	"strconv"
	"strings"

	"gorono/internal/models"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type gauge = models.Gauge
type counter = models.Counter

// три доп. метрики из gopsutil
func GetMoreMetrix() *[]models.Metrics {
	v, _ := mem.VirtualMemory() //             VirtualMemory()
	cc, _ := cpu.Counts(true)
	gaugeMap := map[string]gauge{
		"TotalMemory":     gauge(v.Total),
		"FreeMemory":      gauge(v.Free),
		"CPUutilization1": gauge(cc),
	}
	metrArray := make([]models.Metrics, 0, len(gaugeMap))
	for metrName, metrValue := range gaugeMap {
		mval := float64(metrValue)
		metr := models.Metrics{ID: metrName, MType: "gauge", Value: &mval}
		metrArray = append(metrArray, metr)
	}
	return &metrArray
}

// рантайм метрики из runtime.MemStats
func GetMetrixFromOS() *[]models.Metrics {
	var mS runtime.MemStats
	runtime.ReadMemStats(&mS)
	gaugeMap := map[string]gauge{
		"Alloc":         gauge(mS.Alloc),
		"BuckHashSys":   gauge(mS.BuckHashSys),
		"Frees":         gauge(mS.Frees),
		"GCCPUFraction": gauge(mS.GCCPUFraction),
		"GCSys":         gauge(mS.GCSys),
		"HeapAlloc":     gauge(mS.HeapAlloc),
		"HeapIdle":      gauge(mS.HeapIdle),
		"HeapInuse":     gauge(mS.HeapInuse),
		"HeapObjects":   gauge(mS.HeapObjects),
		"HeapReleased":  gauge(mS.HeapReleased),
		"HeapSys":       gauge(mS.HeapSys),
		"LastGC":        gauge(mS.LastGC),
		"Lookups":       gauge(mS.Lookups),
		"MCacheInuse":   gauge(mS.MCacheInuse),
		"MCacheSys":     gauge(mS.MCacheSys),
		"MSpanInuse":    gauge(mS.MSpanInuse),
		"MSpanSys":      gauge(mS.MSpanSys),
		"Mallocs":       gauge(mS.Mallocs),
		"NextGC":        gauge(mS.NextGC),
		"NumForcedGC":   gauge(mS.NumForcedGC),
		"NumGC":         gauge(mS.NumGC),
		"OtherSys":      gauge(mS.OtherSys),
		"PauseTotalNs":  gauge(mS.PauseTotalNs),
		"StackInuse":    gauge(mS.StackInuse),
		"StackSys":      gauge(mS.StackSys),
		"Sys":           gauge(mS.Sys),
		"TotalAlloc":    gauge(mS.TotalAlloc),
		"RandomValue":   gauge(rand.Float64()), // self-defined
	}
	counterMap := map[string]counter{
		"PollCount": counter(0), // self-defined
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
