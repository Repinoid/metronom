package main

import (
	"context"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func BadPost(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	rwr.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
}

func GetAllMetrix(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	if req.URL.Path != "/" { // if GET with wrong arguments structure
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

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

func GetMetric(rwr http.ResponseWriter, req *http.Request) {
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

func PutMetric(rwr http.ResponseWriter, req *http.Request) {

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
		_ = inter.SaveMS(fileStorePath)
	}
}

func DBPinger(rwr http.ResponseWriter, req *http.Request) {

	//db, err := sql.Open("pgx", dbEndPoint)

	ctx := context.Background()
	db, err := pgx.Connect(ctx, dbEndPoint)

	//	log.Printf("Endpoint is %s\n", dbEndPoint)

	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		//		log.Printf("Open DB error is %v\n", err)
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		return
	}
	defer db.Close(ctx)

	err = db.Ping(ctx)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
		//	log.Printf("PING DB error is %v\n", err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	//log.Printf("AFTER PING DB error is %v\n", err)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}

func DBPingera(rwr http.ResponseWriter, req *http.Request) {
	startt := time.Now()
	defer func(t time.Time) { fmt.Printf("defer Ping time %v µs\n", time.Since(startt).Microseconds()) }(startt)

	err := inter.Ping(ctx)
	sugar.Debugf("Ping time %v µs-----------Inter is %s\n", time.Since(startt).Microseconds(), inter.GetName())
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}
