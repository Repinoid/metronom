package handlera

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
)

// BadPost - Вызывается при неверном URL метода POST
// router.HandleFunc("/", handlera.BadPost).Methods("POST")
func BadPost(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	rwr.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
}

// GetAllMetricsHandler - Выводит на экран все метрики
// router.HandleFunc("/", handlera.GetAllMetricsHandler).Methods("GET")
func GetAllMetricsHandler(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	if req.URL.Path != "/" { // if GET with wrong arguments structure
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metras := []models.Metrics{}
	err := basis.RetryMetricWrapper(models.Inter.GetAllMetrics)(req.Context(), nil, &metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	rwr.WriteHeader(http.StatusOK)

	for _, metr := range metras {
		switch metr.MType {
		case "gauge":
			flo := strconv.FormatFloat(float64(*metr.Value), 'f', -1, 64) // -1 - to remove zeroes tail
			fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%s\n", metr.ID, flo)
		case "counter":
			fmt.Fprintf(rwr, "Counter Metric name   %20s\t\tvalue\t%d\n", metr.ID, *metr.Delta)
		}
	}
}

// GetMetric - Получить значение метрики методом GET
// router.HandleFunc("/value/{metricType}/{metricName}", handlera.GetMetric).Methods("GET")
func GetMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(req)
	metricType := vars["metricType"]
	metricName := vars["metricName"]
	metr := models.Metrics{ID: metricName, MType: metricType}
	err := basis.RetryMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
	if err != nil || !memos.IsMetricOK(metr) { // if no such metric, type+name
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
		// default:
		// 	rwr.WriteHeader(http.StatusNotFound)
		// 	fmt.Fprintf(rwr, `{"wrong metric type":"%s"}`, metricType)
		// 	return
	}
}

// PutMetric - Внести метрику методом GET
// router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlera.PutMetric).Methods("POST")
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
	basis.RetryMetricWrapper(models.Inter.PutMetric)(req.Context(), &metr, nil)
	err := basis.RetryMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
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
	if models.StoreInterval == 0 {
		_ = models.Inter.SaveMS(models.FileStorePath)
	}
}

// DBPinger - Пинг базы данных
// router.HandleFunc("/ping", handlera.DBPinger).Methods("GET")
func DBPinger(rwr http.ResponseWriter, req *http.Request) {

	err := models.Inter.Ping(req.Context(), models.DBEndPoint)
	if err != nil {
		rwr.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)
}
