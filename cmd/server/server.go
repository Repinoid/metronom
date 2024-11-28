package main

import (
	//"fmt"
	"fmt"
	"net/http"
	"strconv"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau   map[string]gauge
	count map[string]counter
}

const localPort = ":8080"

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
func (ms *MemStorage) getCounterValue(name string, value *string) int {
	if _, ok := ms.count[name]; ok {
		*value = strconv.FormatInt(int64(ms.count[name]), 10)
		return http.StatusOK
	}
	return http.StatusNotFound
}
func (ms *MemStorage) getGaugeValue(name string, value *string) int {
	if _, ok := ms.gau[name]; ok {
		//		*value = strconv.FormatFloat(float64(ms.gau[name]), 'f', 4, 64)
		*value = fmt.Sprintf("%f", ms.gau[name])
		return http.StatusOK
	}
	return http.StatusNotFound
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

	router := http.NewServeMux()
	router.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", treatMetric)
	router.HandleFunc("GET /value/{metricType}/{metricName}", getMetric)
	router.HandleFunc("GET /", getAllMetrix)
	router.HandleFunc("POST /", badPost)

	return http.ListenAndServe(localPort, router)
}

func badPost(rwr http.ResponseWriter, req *http.Request) {
	rwr.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rwr, "POST http.StatusNotFound with %s\n", req.URL.Path)
}

func getAllMetrix(rwr http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, "BadRequest with %s\n", req.URL.Path)
		return
	}

	for nam, val := range memStor.gau {
		fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%6.4f\n", nam, val)
	}
	for nam, val := range memStor.count {
		fmt.Fprintf(rwr, "Counter Metric name %20s\t\tvalue\t%d\n", nam, val)
	}
	rwr.WriteHeader(http.StatusOK)
}
func getMetric(rwr http.ResponseWriter, req *http.Request) {
	val := "badly"
	status := http.StatusBadRequest
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	if metricType == "gauge" {
		status = memStor.getGaugeValue(metricName, &val)
	}
	if metricType == "counter" {
		status = memStor.getCounterValue(metricName, &val)
	}
	if status == http.StatusOK {
		rwr.WriteHeader(http.StatusOK)
		fmt.Fprint(rwr, val)
	} else {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, "BadRequest, No value for %s of %s type\n", metricName, metricType)
	}

}

func treatMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")
	if metricValue == "" {
		rwr.WriteHeader(http.StatusNotFound)
		return
	}

	if metricType != "gauge" && metricType != "counter" {
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
