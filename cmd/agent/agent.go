package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau    map[string]gauge
	count  map[string]counter
	mutter sync.RWMutex
	//	PollCount int
}

// var memStor *MemStorage
var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2

func getMetrix(memStor *MemStorage) error {
	var mS runtime.MemStats
	runtime.ReadMemStats(&mS)
	memStor.mutter.Lock() // MUTEXed
	defer memStor.mutter.Unlock()

	//	memStor.PollCount++
	memStor.gau = map[string]gauge{
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
	memStor.count = map[string]counter{
		"PollCount": counter(0), // self-defined
	}
	return nil
}
func postMetric(metricType, metricName, metricValue string) error {
	url := "http://" + host + "/update/" + metricType + "/" + metricName + "/" + metricValue
	resp, err := http.Post(url, "text/plain", nil)
	if err == nil {
		defer resp.Body.Close()
	}
	return err
}

func main() {
	if err := foa4Agent(); err != nil {
		log.Println(err, " no success for foa4Agent() ")
		return
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	memStor := new(MemStorage)
	for {
		cunt := 0
		for i := 0; i < reportInterval/pollInterval; i++ {
			err := getMetrix(memStor)
			if err != nil {
				log.Println(err, "getMetrix")
			} else {
				cunt++
			}
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
		for name, value := range memStor.gau {
			valStr := strconv.FormatFloat(float64(value), 'f', 4, 64)
			err := postMetric("gauge", name, valStr)
			if err != nil {
				log.Println(err, "gauge", name, valStr)
			}
		}
		for name := range memStor.count {
			valStr := strconv.FormatInt(int64(cunt), 10)
			err := postMetric("counter", name, valStr)
			if err != nil {
				log.Println(err, "counter", name, valStr)
			}
		}
	}
}
