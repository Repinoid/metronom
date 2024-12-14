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
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	switch metricType {
	case "counter":
		var cunt counter
		if memStor.getCounterValue(metricName, &cunt) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		fmt.Fprint(rwr, cunt)
	case "gauge":
		var gaaga gauge
		if memStor.getGaugeValue(metricName, &gaaga) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		fmt.Fprint(rwr, gaaga)
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(rwr, nil)
		return
	}
}

func treatMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(req)
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metricValue := vars["metricValue"]
	if metricValue == "" {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addCounter(metricName, counter(value))
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addGauge(metricName, gauge(value))
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}

// metricstest -test.v -test.run="^TestIteration6[AB]*$" -binary-path=cmd/server/server.exe -source-path=cmd/server/
