// сервер для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"

	"gorono/internal/handlera"
	"gorono/internal/middlas"
	"gorono/internal/models"
)

// listens on the TCP network address for ListenAndServe
//var Host = "localhost:8000"

var Host = ":8080"

// Глобальные переменные для флага компилляции.
// Форма запуска go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X main.buildCommit=comitta" main.go
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	if err := InitServer(); err != nil {
		log.Println(err, " no success for foa4Server() ")
		return
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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

	//router.HandleFunc("/s", seconda).Methods("GET")

	router.Use(middlas.GzipHandleEncoder)
	router.Use(middlas.GzipHandleDecoder)
	//router.Use(middlas.NoSugarLogging)	// или NoSugarLogging - или WithLogging ZAP логирование
	router.Use(middlas.WithLogging)
	router.Use(middlas.CryptoHandleDecoder)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	return http.ListenAndServe(Host, router)
}

// func seconda(rwr http.ResponseWriter, req *http.Request) {
// 	//	var DBEndPoint = "postgres://postgres:postgres@go_db:5432/postgres"

// 	//	baza, err := pgx.Connect(context.Background(), DBEndPoint)
// 	baza, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_DSN"))
// 	if err != nil {
// 		fmt.Fprintf(rwr, "NO pgx.Connect %v\n", err)
// 		return
// 	}
// 	fmt.Fprintf(rwr, "PING OK %v %v\n", baza, time.Now())

// }
