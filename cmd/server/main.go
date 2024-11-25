package main

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau   map[string]gauge
	count map[string]counter
}

func (ms *MemStorage) initMemStorage() error {
	ms.gau = make(map[string]gauge)
	ms.count = make(map[string]counter)
	return nil
}
func (ms *MemStorage) addGauge(name string, value gauge) error {
	ms.gau[name] = value
	return nil
}
func (ms *MemStorage) addCounter(name string, value counter) error {
	if _, ok := ms.count[name]; ok {
		ms.count[name] += value
		return nil
	}
	ms.count[name] = value
	return nil
}

var memStor *MemStorage

func main() {
	memStor = new(MemStorage)
	memStor.initMemStorage()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(webhook))
}

func webhook(rwr http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // разрешаем только POST-запросы
		rwr.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rwr.Header().Set("Content-Type", "text/plain")

	urla := r.URL.String()
	splittedURL := strings.Split(urla, "/")

	if len(splittedURL) < 5 {
		rwr.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := splittedURL[2]
	metricName := splittedURL[3]
	metricValue := splittedURL[4]
	if splittedURL[1] != "update" || (metricType != "gauge" && metricType != "counter") {
		rwr.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			return
		}
		memStor.addCounter(metricName, counter(value))
	} else { //	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			return
		}
		memStor.addGauge(metricName, gauge(value))
	}
	rwr.WriteHeader(http.StatusOK)
}
