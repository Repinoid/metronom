// сервер для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	if err := Run(); err != nil {
		log.Printf("Server Shutdown by syscall, ListenAndServe message -  %v\n", err)
	}
}

// run. Запуск сервера и хендлеры
func Run() (err error) {

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
	//router.Use(middlas.NoSugarLogging)	// или NoSugarLogging - или WithLogging ZAP логирование
	router.Use(middlas.WithLogging)
	router.Use(middlas.CryptoHandleDecoder)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	//
	var srv = http.Server{Addr: Host, Handler: router}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-exit
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() error {
		return srv.ListenAndServe()
	}()

	go func() error {
		<-ctx.Done()
		wg.Done()
		return srv.Shutdown(ctx)
	}()

	// запись метрик в файл
	if models.StoreInterval > 0 {
		wg.Add(1)
		go models.Inter.Saver(ctx, models.FileStorePath, models.StoreInterval, &wg)
	}

	wg.Wait()

	models.Inter.Close()

	log.Println("Server Shutdown gracefully")

	return err
}
