// сервер для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	pb "gorono/cmd/proto"

	"gorono/internal/gremote"
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
	router.Use(middlas.IpcidrCheck)

	router.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	//

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-exit
		cancel()
	}()

	var wg sync.WaitGroup

	// HTTP server gouroutine launch
	var httpServer = http.Server{Addr: Host, Handler: router}
	wg.Add(1)
	go func() {

		// ListenAndServe always returns a non-nil error.
		// After Server.Shutdown or Server.Close, the returned error is ErrServerClosed. иначе фатал
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// gRPC server, default port ":3200" or from parameters
	listen, err := net.Listen("tcp", models.Gport)
	if err != nil {
		log.Fatal(err)
	}

	// серверу для шифрования необходимы оба сертификата, приватный и публичный
	// for testing purposes use own certificates - потому что в задании без grpc для сервера задаётся только приватный сертификат
	var grpcServer *grpc.Server
	//	if isCoded {
	// Load TLS credentials
	creds, err := LoadTLSCredentials("../tls/cert.pem", "../tls/key.pem")
	if err != nil {
		log.Fatalf("failed to load TLS credentials: %v", err)
	}
	grpcServer = grpc.NewServer(grpc.UnaryInterceptor(middlas.OurInterceptor), grpc.Creds(creds))
	//} else {
	// без шифровки
	//grpcServer = grpc.NewServer()
	//}

	// регистрируем сервис
	pb.RegisterMetricServer(grpcServer, &gremote.MetricServer{})

	// gRPC server gouroutine launch
	wg.Add(1)
	go func() {
		fmt.Println("Сервер gRPC начал работу")
		// получаем запрос gRPC
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()

	// graceful shutdown when cancel() of context sends Done
	go func() {
		<-ctx.Done()
		defer wg.Done() // for http server
		defer wg.Done() // for grpc server
		// Attempt graceful shutdown
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := httpServer.Close(); err != nil {
				log.Fatalf("Forced shutdown failed: %v", err)
			}
		} else {
			defer fmt.Println("Сервер HTTP Shutdown gracefully")
		}
		// GracefulStop stops the gRPC server gracefully.
		// It stops the server from accepting new connections and RPCs and blocks until all the pending RPCs are finished.
		grpcServer.GracefulStop()
		// defer для порядку
		defer fmt.Println("Сервер gRPC Shutdown gracefully")
	}()

	// запись метрик в файл
	if models.StoreInterval > 0 {
		wg.Add(1)
		go models.Inter.Saver(ctx, models.FileStorePath, models.StoreInterval, &wg)
	}

	wg.Wait()

	models.Inter.Close()

	log.Println("All services Shutdown gracefully")

	return err
}
