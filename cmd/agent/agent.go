package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("wrong counter value %w", err)
		}
		metr = Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &val,
		}
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("wrong gauge value %w", err)
		}
		metr = Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &val,
		}
	default:
		return fmt.Errorf("wrong metric type")
	}
	march, err := json.Marshal(metr)
	if err != nil {
		return fmt.Errorf("could not marshal metr %w", err)
	}
	resp, err := http.Post("http://"+host+"/update/", "application/json", bytes.NewBuffer(march))
	if err != nil {
		return fmt.Errorf("could not post %w", err)
	}
	defer resp.Body.Close()

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
	memStor := MemStorage{} //memStor := new(MemStorage)
	for {
		cunt := 0
		for i := 0; i < reportInterval/pollInterval; i++ {
			err := getMetrix(&memStor)
			if err != nil {
				log.Println(err, "getMetrix")
			} else {
				cunt++
			}
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}

		memStor.count["PollCount"] = counter(cunt)
		bunch := makeBunchOfMetrics(&memStor)
		log.Println(len(bunch))

		err := postBunch(bunch)
		if err != nil {
			log.Printf("AGENT postBunch ERROR %+v\n", err)
		}

		// for name, value := range memStor.gau {
		// 	valStr := strconv.FormatFloat(float64(value), 'f', 4, 64)
		// 	err := postMetric("gauge", name, valStr)
		// 	if err != nil {
		// 		log.Println(err, "gauge", name, valStr)
		// 	}
		// }
		// for name := range memStor.count {
		// 	valStr := strconv.FormatInt(int64(cunt), 10)
		// 	err := postMetric("counter", name, valStr)
		// 	if err != nil {
		// 		log.Println(err, "counter", name, valStr)
		// 	}
		// }

	}
}

func postBunch(bunch []Metrics) error {
	marshalledBunch, err := json.Marshal(bunch)
	if err != nil {
		return err
	}
	compressedBunch, err := pack2gzip(marshalledBunch)
	if err != nil {
		return err
	}
	httpc := resty.New() //
	httpc.SetBaseURL("http://" + host)

	httpc.SetRetryCount(3)
	httpc.SetRetryWaitTime(1 * time.Second)    // начальное время повтора
	httpc.SetRetryMaxWaitTime(9 * time.Second) // 1+3+5
	//tn := time.Now()                           // -------------
	httpc.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		rwt := client.RetryWaitTime
		//	fmt.Printf("waittime \t%+v\t time %+v  count %d\n", rwt, time.Since(tn), client.RetryCount) // -------
		client.SetRetryWaitTime(rwt + 2*time.Second)
		//	tn = time.Now() // ----------------
		return client.RetryWaitTime, nil
	})

	req := httpc.R().
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedBunch).
		SetHeader("Accept-Encoding", "gzip")

	_, err = req.
		SetDoNotParseResponse(false).
		Post("/updates/")

		//	log.Printf("%+v\n", resp)

	return err
}

func makeBunchOfMetrics(memStor *MemStorage) []Metrics {
	metrArray := make([]Metrics, 0, len(memStor.gau)+len(memStor.count))

	for metrName, metrValue := range memStor.count {
		mval := int64(metrValue)
		metr := Metrics{ID: metrName, MType: "counter", Delta: &mval}
		metrArray = append(metrArray, metr)
	}
	for metrName, metrValue := range memStor.gau {
		mval := float64(metrValue)
		metr := Metrics{ID: metrName, MType: "gauge", Value: &mval}
		metrArray = append(metrArray, metr)
	}
	return metrArray
}

func pack2gzip(data2pack []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	//	zw.ModTime = time.Now()
	_, err := zw.Write(data2pack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Write %w ", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Close %w ", err)
	}
	return buf.Bytes(), nil
}
