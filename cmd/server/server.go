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
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"internal/dbaser"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type gauge float64
type counter int64
type MemStorage struct {
	gau    map[string]gauge
	count  map[string]counter
	mutter sync.RWMutex
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var memStor MemStorage
var host = "localhost:8080"
var sugar zap.SugaredLogger
var isBase = false

type str4db struct {
	MetricBase *pgx.Conn
}

var MetricBase str4db

var check *bool

func saver(memStor *MemStorage, fnam string) error {

	for {
		time.Sleep(time.Duration(storeInterval) * time.Second)
		err := memStor.SaveMS(fnam)
		if err != nil {
			return fmt.Errorf("save err %v", err)
		}
	}
}

func main() {
	if err := foa4Server(); err != nil {
		log.Println(err, " no success for foa4Server() ")
		return
	}

	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}

	if reStore {
		_ = memStor.LoadMS(fileStorePath)
	}

	if storeInterval > 0 {
		go saver(&memStor, fileStorePath)
	}

	ctx := context.Background()
	mb, err := pgx.Connect(ctx, dbEndPoint)
	MetricBase = str4db{MetricBase: mb}
	if err != nil {
		isBase = false
		log.Printf("Can't connect to DB %s\n", dbEndPoint)
	} else {
		err = dbaser.TableCreation(ctx, MetricBase.MetricBase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create tables: %v\n", err)
		} else {
			isBase = true
		}
	}
	tr := true
	//	log.Printf("%v\nisBase - %v\ncheck - %v\n\n\n", MetricBase.MetricBase.Config().Host, isBase, check)
	check = &tr

	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", WithLogging(treatMetric)).Methods("POST")
	router.HandleFunc("/update/", WithLogging(treatJSONMetric)).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", WithLogging(getMetric)).Methods("GET")
	router.HandleFunc("/value/", WithLogging(getJSONMetric)).Methods("POST")
	router.HandleFunc("/", WithLogging(getAllMetrix)).Methods("GET")
	router.HandleFunc("/", WithLogging(badPost)).Methods("POST") // if POST with wrong arguments structure
	//	router.HandleFunc("/ping", WithLogging(dbPinger)).Methods("GET")
	router.HandleFunc("/ping", dbPinger).Methods("GET")

	router.Use(gzipHandleEncoder)
	router.Use(gzipHandleDecoder)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	return http.ListenAndServe(host, router)
}

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
	memStor.mutter.RLock() // <---- MUTEX
	defer memStor.mutter.RUnlock()
	for nam, val := range memStor.gau {
		flo := strconv.FormatFloat(float64(val), 'f', -1, 64) // -1 - to remove zeroes tail
		fmt.Fprintf(rwr, "Gauge Metric name   %20s\t\tvalue\t%s\n", nam, flo)
	}
	for nam, val := range memStor.count {
		fmt.Fprintf(rwr, "Counter Metric name %20s\t\tvalue\t%d\n", nam, val)
	}
	rwr.WriteHeader(http.StatusOK)
}
func getMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "text/html")
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
	rwr.WriteHeader(http.StatusOK)
}

func treatMetric(rwr http.ResponseWriter, req *http.Request) {

	//log.Printf("%v\nisBase - %v\ncheck - %v\n\n\n", MetricBase.MetricBase, isBase, check)

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
	rwr.WriteHeader(http.StatusOK)
	fmt.Fprintf(rwr, `{"status":"StatusOK"}`)

	if storeInterval == 0 {
		_ = memStor.SaveMS(fileStorePath)
	}
}

func dbPinger(rwr http.ResponseWriter, req *http.Request) {

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

/*
metricstest -test.v -test.run="^TestIteration11[AB]*$" ^
-binary-path=cmd/server/server.exe -source-path=cmd/server/ ^
-agent-binary-path=cmd/agent/agent.exe ^
-server-port=8080 -file-storage-path=goshran.txt ^
-database-dsn=postgres://postgres:passwordas@localhost:5432/postgres


metricstest -test.v -test.run="^TestIteration1[AB]*$" -binary-path=cmd/server/server.exe -source-path=cmd/server/

go run . -d=postgres://postgres:passwordas@localhost:5432/postgres

*/
