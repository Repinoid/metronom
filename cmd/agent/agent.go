package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau   map[string]gauge
	count map[string]counter
}

var memStor *MemStorage
var PollCount int
var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2

func getMetrix(memStor *MemStorage) error {
	var mS runtime.MemStats
	runtime.ReadMemStats(&mS)
	PollCount++
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
		"RandomValue":   gauge(rand.Float64()),
	}
	memStor.count = map[string]counter{
		"PollCount": counter(PollCount),
	}
	return nil
}
func postMetric(metricType, metricName, metricValue string) int {

	url := "http://" + host + "/update/" + metricType + "/" + metricName + "/" + metricValue
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return resp.StatusCode
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
func main() {
	if useClientArguments() != 0 {
		return
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	memStor = new(MemStorage)

	for {
		for i := 0; i < reportInterval/pollInterval; i++ {
			err := getMetrix(memStor)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
		for name, value := range memStor.gau {
			status := postMetric("gauge", name, strconv.FormatFloat(float64(value), 'f', 4, 64))
			if status != http.StatusOK {
				log.Println(status)
			}
		}
		for name, value := range memStor.count {
			status := postMetric("counter", name, strconv.FormatInt(int64(value), 10))
			if status != http.StatusOK {
				log.Println(status)
			}
		}
	}
}

func useClientArguments() int {
	args := os.Args[1:]

	for _, a := range args {
		if len(a) < 3 {
			fmt.Printf("unknown Argument -  %s\n", a)
			return 1
		}
		flagus := a[:3]
		tail := a[3:]
		switch flagus {
		case "-a=":
			host = a[3:]
		case "-r=":
			secs, err := strconv.Atoi(tail)
			if err != nil {
				fmt.Printf("Bad Argument for reportInterval with %s\n", tail)
				return 2
			}
			reportInterval = secs
		case "-p=":
			secs, err := strconv.Atoi(tail)
			if err != nil {
				fmt.Printf("Bad Argument for pollInterval with %s\n", tail)
				return 3
			}
			pollInterval = secs
		default:
			fmt.Printf("unknown Argument -  %s\n", a)
			return 4
		}
	}
	return 0
}
