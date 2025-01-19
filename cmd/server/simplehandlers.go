package main

import (
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/models"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

func badPost(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	rwr.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
}

func getAllMetrix(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	if req.URL.Path != "/" { // if GET with wrong arguments structure
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	var mutter sync.RWMutex
	mutter.RLock() // <---- MUTEX
	defer mutter.RUnlock()

	metras, err := basis.GetAllMetricsWrapper(inter.GetAllMetrics)(ctx)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	rwr.WriteHeader(http.StatusOK)
	for _, metr := range *metras {
		switch metr.MType {
		case "gauge":
			flo := strconv.FormatFloat(float64(*metr.Value), 'f', -1, 64) // -1 - to remove zeroes tail
			fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%s\n", metr.ID, flo)
		case "counter":
			fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%d\n", metr.ID, *metr.Delta)
		}
	}
}

func getMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(req)
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metr := models.Metrics{ID: metricName, MType: metricType}
	metr, err := basis.GetMetricWrapper(inter.GetMetric)(ctx, &metr) //inter.GetMetric(ctx, &metr)
	if err != nil || !models.IsMetricsOK(metr) {                     // if no such metric, type+name
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"wrong metric name":"%s"}`, metricName)
		return
	}
	switch metricType {
	case "gauge":
		rwr.WriteHeader(http.StatusOK)
		fmt.Fprint(rwr, *metr.Value)
	case "counter":
		rwr.WriteHeader(http.StatusOK)
		fmt.Fprint(rwr, *metr.Delta)
	default:
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"wrong metric type":"%s"}`, metricType)
		return
	}
}

func putMetric(rwr http.ResponseWriter, req *http.Request) {

	rwr.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(req)
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metricValue := vars["metricValue"]
	if metricValue == "" {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	metr := models.Metrics{}
	switch metricType {
	case "counter":
		out, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		metr = models.Metrics{ID: metricName, MType: "counter", Delta: &out}
	//	basis.PutMetricWrapper(inter.PutMetric)(ctx, &metr) //inter.PutMetric(ctx, &metr)
	case "gauge":
		out, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		metr = models.Metrics{ID: metricName, MType: "gauge", Value: &out}
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	basis.PutMetricWrapper(inter.PutMetric)(ctx, &metr)              //inter.PutMetric(ctx, &metr)
	metr, err := basis.GetMetricWrapper(inter.GetMetric)(ctx, &metr) // inter.GetMetric(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	switch metr.MType {
	case "gauge":
		fmt.Fprint(rwr, *metr.Value)
	case "counter":
		fmt.Fprint(rwr, *metr.Delta)
	}
	if storeInterval == 0 {
		_ = memStor.SaveMS(fileStorePath)
	}
}

func dbPinger(rwr http.ResponseWriter, req *http.Request) {

	err := inter.Ping(ctx)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}
