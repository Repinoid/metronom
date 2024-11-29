package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau   map[string]gauge
	count map[string]counter
}

var host = "localhost:8080"

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
		*value = strconv.FormatFloat(float64(ms.gau[name]), 'f', -1, 64)
		return http.StatusOK
	}
	return http.StatusNotFound
}

var memStor *MemStorage

func main() {
	if useServerArguments() != 0 {
		return
	}

	memStor = new(MemStorage)
	memStor.initMemStorage()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", treatMetric).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", getMetric).Methods("GET")
	router.HandleFunc("/", getAllMetrix).Methods("GET")
	router.HandleFunc("/", badPost).Methods("POST")

	return http.ListenAndServe(host, router)
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
		flo := strconv.FormatFloat(float64(val), 'f', -1, 64)
		fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%s\n", nam, flo)
	}
	for nam, val := range memStor.count {
		fmt.Fprintf(rwr, "Counter Metric name %20s\t\tvalue\t%d\n", nam, val)
	}
	rwr.WriteHeader(http.StatusOK)
}
func getMetric(rwr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	val := "badly"
	status := http.StatusNotFound
	metricType := vars["metricType"]
	metricName := vars["metricName"]
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
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, "BadRequest, No value for %s of %s type\n", metricName, metricType)
	}

}

func treatMetric(rwr http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	rwr.Header().Set("Content-Type", "text/plain")
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metricValue := vars["metricValue"]
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

func useServerArguments() int {
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
		return 0
	}
	args := os.Args[1:]
	for _, a := range args {
		if len(a) < 3 {
			fmt.Printf("unknown Argument -  %s\n", a)
			return 1
		}
		flagus := a[:3]
		switch flagus {
		case "-a=":
			host = a[3:]
		default:
			fmt.Printf("unknown Argument -  %s\n", a)
			return 4
		}
	}
	return 0
}
