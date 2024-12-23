package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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
	var metr Metrics
	switch metricType {
	case "counter":
		val, _ := strconv.ParseInt(metricValue, 10, 64)
		metr = Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &val,
		}
		march, _ := json.Marshal(metr)
		resp, err := http.Post("http://"+host+"/update/", "application/json", bytes.NewBuffer(march))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	case "gauge":
		val, _ := strconv.ParseFloat(metricValue, 64)
		metr = Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &val,
		}
		march, _ := json.Marshal(metr)
		resp, err := http.Post("http://"+host+"/update/", "application/json", bytes.NewBuffer(march))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	default:
		return fmt.Errorf("wrong metric type")
	}
	return nil
}

func main() {
	if err := foa4Agent(); err != nil {
		log.Fatal("INTERVAL error ", err)
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
