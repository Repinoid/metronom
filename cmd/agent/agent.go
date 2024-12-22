package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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

func pack2gzip(data2pack []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.ModTime = time.Now()
	_, err := zw.Write(data2pack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Write %w ", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Close %w ", err)
	}
	return buf.Bytes(), nil
}
func unpackFromGzip(data2unpack io.Reader) (io.Reader, error) {
	gzipReader, err := gzip.NewReader(data2unpack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader %w ", err)
	}
	if err := gzipReader.Close(); err != nil {
		return nil, fmt.Errorf("zr.Close %w ", err)
	}
	return gzipReader, nil
}

func postByNewRequest(metr Metrics) ([]byte, error) {
	jsonStrMarshalled, err := json.Marshal(metr)
	if err != nil {
		return nil, fmt.Errorf("marshal err %w ", err)
	}
	jsonStrPacked, err := pack2gzip(jsonStrMarshalled)
	if err != nil {
		return nil, fmt.Errorf("pack2gzip %w ", err)
	}
	requerest, err := http.NewRequest("POST", "http://"+host+"/update/", bytes.NewBuffer(jsonStrPacked))
	if err != nil {
		return nil, fmt.Errorf("erra http.NewRequest %w ", err)
	}
	requerest.Header.Set("Accept-Encoding", "gzip")
	requerest.Header.Set("Content-Encoding", "gzip")
<<<<<<< HEAD
	requerest.Header.Set("Content-Type", "application/json")
=======
>>>>>>> 2f3e239640b9d392f64b81867ca4d4ae40c08a36

	client := &http.Client{}

	responsa, err := client.Do(requerest)
	if err != nil {
		return nil, fmt.Errorf("client.Do  %w ", err)
	}
	defer responsa.Body.Close()
<<<<<<< HEAD
	
=======
>>>>>>> 2f3e239640b9d392f64b81867ca4d4ae40c08a36
	var reader io.Reader
	if responsa.Header.Get(`Content-Encoding`) == `gzip` {
		reader, err = unpackFromGzip(responsa.Body)
		if err != nil {
			return nil, fmt.Errorf("unpackFromGzip %w ", err)
		}
	} else {
		reader = responsa.Body
	}
	telo, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll(reader) %w ", err)
	}
	return telo, nil

}

func postMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("strconv.ParseInt(metricValue, 10, 64) %w", err)
		}
		metr := Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &val,
		}
		_, err = postByNewRequest(metr)
		if err != nil {
			return fmt.Errorf("postByNewRequest counter %w", err)
		}
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("strconv.ParseFloat(metricValue, 64) %w", err)
		}
		metr := Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &val,
		}
		_, err = postByNewRequest(metr)
		if err != nil {
			return fmt.Errorf("postByNewRequest gauge %w", err)
		}
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
