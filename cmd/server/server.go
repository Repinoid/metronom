// сервер для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.
package main

import (
	"log"
	"net/http"

	"gorono/internal/handlera"
	"gorono/internal/middlas"
	"gorono/internal/models"

	_ "net/http/pprof"

	"github.com/gorilla/mux"
)

// listens on the TCP network address for ListenAndServe
var Host = "localhost:8080"

//var ctx context.Context

func main() {

	if err := InitServer(); err != nil {
		log.Println(err, " no success for foa4Server() ")
		return
	}

	if models.ReStore {
		_ = models.Inter.LoadMS(models.FileStorePath)
	}

	if models.StoreInterval > 0 {
		go models.Inter.Saver(models.FileStorePath, models.StoreInterval)
	}

	if err := Run(); err != nil {
		panic(err)
	}
}

// run. ЗАпуск сервера и хендлеры
func Run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlera.PutMetric).Methods("POST")
	router.HandleFunc("/update/", handlera.PutJSONMetric).Methods("POST")
	router.HandleFunc("/updates/", handlera.Buncheras).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", handlera.GetMetric).Methods("GET")
	router.HandleFunc("/value/", handlera.GetJSONMetric).Methods("POST")
	router.HandleFunc("/", handlera.GetAllMetricsHandler).Methods("GET")
	router.HandleFunc("/", handlera.BadPost).Methods("POST") // if POST with wrong arguments structure
	router.HandleFunc("/ping", handlera.DBPinger).Methods("GET")

	router.Use(middlas.GzipHandleEncoder)
	router.Use(middlas.GzipHandleDecoder)
	//router.Use(middlas.NoSugarLogging)
	router.Use(middlas.WithLogging)
	router.Use(middlas.CryptoHandleDecoder)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	return http.ListenAndServe(Host, router)
}

/*
go test ./... -v -coverpkg=./...

http://localhost:8080/pkg/?m=all

godoc -http=:8080 -play

*/
