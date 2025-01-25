/*
metricstest -test.v -test.run="^TestIteration10[AB]*$" ^
-binary-path=cmd/server/server.exe -source-path=cmd/server/ ^
-agent-binary-path=cmd/agent/agent.exe ^
-server-port=8080 -file-storage-path=goshran.txt ^
-database-dsn=postgres://postgres:passwordas@localhost:5432/postgres


curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"nam\",\"value\":77}"
*/

package main

import (
	"context"
	"log"
	"net/http"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Metrics = memos.Metrics
type MemStorage = memos.MemoryStorageStruct

var host = "localhost:8080"
var sugar zap.SugaredLogger

var ctx context.Context

var inter models.Inter // 	= memStor OR dbStorage

func main() {

	if err := InitServer(); err != nil {
		log.Println(err, " no success for foa4Server() ")
		return
	}

	if reStore {
		_ = inter.LoadMS(fileStorePath)
	}

	if storeInterval > 0 {
		go inter.Saver(fileStorePath, storeInterval)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", PutMetric).Methods("POST")
	router.HandleFunc("/update/", PutJSONMetric).Methods("POST")
	router.HandleFunc("/updates/", Buncheras).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", GetMetric).Methods("GET")
	router.HandleFunc("/value/", GetJSONMetric).Methods("POST")
	router.HandleFunc("/", GetAllMetrix).Methods("GET")
	router.HandleFunc("/", BadPost).Methods("POST") // if POST with wrong arguments structure
	router.HandleFunc("/ping", DBPinger).Methods("GET")

	router.Use(middlas.GzipHandleEncoder)
	router.Use(middlas.GzipHandleDecoder)
	router.Use(middlas.WithLogging)

	return http.ListenAndServe(host, router)
}

/*
metricstest -test.v -test.run="^TestIteration11[AB]*$" ^
-binary-path=cmd/server/server.exe -source-path=cmd/server/ ^
-agent-binary-path=cmd/agent/agent.exe ^
-server-port=8080 -file-storage-path=goshran.txt ^
-database-dsn=postgres://postgres:passwordas@localhost:5432/postgres


metricstest -test.v -test.run="^TestIteration1[AB]*$" -binary-path=cmd/server/server.exe -source-path=cmd/server/

go run . -d=postgres://postgres:passwordas@localhost:5432/postgres

*/
