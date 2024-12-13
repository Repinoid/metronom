package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau    map[string]gauge
	count  map[string]counter
	mutter sync.RWMutex
}

var memStor MemStorage
var host = "localhost:8080"
var sugar zap.SugaredLogger

func main() {
	if err := foa4Server(); err != nil {
		log.Println(err, " no success for foa4Server() ")
		return
	}

	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", WithLogging(treatMetric)).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", WithLogging(getMetric)).Methods("GET")
	router.HandleFunc("/", WithLogging(getAllMetrix)).Methods("GET")
	router.HandleFunc("/", WithLogging(badPost)).Methods("POST") // if POST with wrong arguments structure

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	return http.ListenAndServe(host, router)
}

func badPost(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	rwr.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
}

func getAllMetrix(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	if req.URL.Path != "/" { // if GET with wrong arguments structure
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	memStor.mutter.RLock() // <---- MUTEX
	defer memStor.mutter.RUnlock()
	for nam, val := range memStor.gau {
		flo := strconv.FormatFloat(float64(val), 'f', -1, 64) // -1 - to remove zeroes tail
		fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%s\n", nam, flo)
	}
	for nam, val := range memStor.count {
		fmt.Fprintf(rwr, "Counter Metric name %20s\t\tvalue\t%d\n", nam, val)
	}
}
func getMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(req)
	val := "badly" // does not matter what initial value, could be "var val string"
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	if metricType != "gauge" && metricType != "counter" {
		rwr.WriteHeader(http.StatusBadRequest)
		return
	}
	var err error
	if metricType == "gauge" {
		err = memStor.getGaugeValue(metricName, &val)
	} else { //if metricType == "counter" {
		err = memStor.getCounterValue(metricName, &val)
	}
	if err == nil {
		rwr.WriteHeader(http.StatusOK)
		fmt.Fprint(rwr, val)
	} else {
		rwr.WriteHeader(http.StatusNotFound)
		log.Printf("BadRequest, No value for %s of %s type", metricName, metricType)
	}
}

func treatMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(req)
	rwr.Header().Set("Content-Type", "text/plain")
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metricValue := vars["metricValue"]
	if metricValue == "" {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	if metricType != "gauge" && metricType != "counter" {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addCounter(metricName, counter(value))
	} else { //	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addGauge(metricName, gauge(value))
	}
	//	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}
