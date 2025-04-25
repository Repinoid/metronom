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

	if models.StoreInterval > 0 {
		go models.Inter.Saver(models.FileStorePath, models.StoreInterval)
	}

	if err := Run(); err != nil {
		panic(err)
	}
}

// run. ЗАпуск сервера и хендлеры
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

	//servak := http.ListenAndServe(Host, router)

	var srv = http.Server{Addr: Host, Handler: router}

	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigint
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		if err := srv.Shutdown(context.Background()); err != nil {
			// ошибки закрытия Listener
			log.Printf("HTTP server Shutdown: %v", err)
		}
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()
	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		// ошибки старта или остановки Listener
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed

	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы
	
	fmt.Println("Server Shutdown gracefully")

	return err
}
