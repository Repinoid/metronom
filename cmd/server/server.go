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
	"time"

	"internal/dbaser"
	"internal/memo"
	"internal/middles"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
)

type gauge = memo.Gauge
type counter = memo.Counter
type MemStorage = memo.MemStorage

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var memStor MemStorage
var host = "localhost:8080"
var sugar zap.SugaredLogger

//var isBase = false

// type str4db struct {
// 	ctx        context.Context
// 	isBase     bool
// 	MetricBase *pgx.Conn
// }

var MetricBaseStruct dbaser.Struct4db

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

	// memStor = MemStorage{}
	// memStor.Gaugemetr = map[string]gauge{}
	// memStor.Countmetr = make(map[string]counter)
	memStor = MemStorage{
		Gaugemetr: make(map[string]gauge),
		Countmetr: make(map[string]counter),
	}

	if reStore  {
		//_ = LoadMS(&memStor, "Y:/GO/ypro/goshran.txt")
		//		_ = memo.LoadMS(&memStor, fileStorePath)
		_ = memStor.LoadMS(fileStorePath)

	}

	//log.Printf("%+v\t%+v\n", memStor.Countmetr, memStor.Gaugemetr)
	//log.Printf("base url %v\t\t\tis connected %v\n\n\n", MetricBaseStruct.MetricBase.Config().Host, MetricBaseStruct.IsBase)

	if storeInterval > 0 {
		go saver(&memStor, fileStorePath)
	}

	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {

	router := mux.NewRouter()
	router.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", middles.WithLogging(treatMetric)).Methods("POST")
	router.HandleFunc("/update/", middles.WithLogging(treatJSONMetric)).Methods("POST")
	router.HandleFunc("/updates/", middles.WithLogging(buncheras)).Methods("POST")
	router.HandleFunc("/value/{metricType}/{metricName}", middles.WithLogging(getMetric)).Methods("GET")
	router.HandleFunc("/value/", middles.WithLogging(getJSONMetric)).Methods("POST")
	router.HandleFunc("/", middles.WithLogging(getAllMetrix)).Methods("GET")
	router.HandleFunc("/", middles.WithLogging(badPost)).Methods("POST") // if POST with wrong arguments structure
	//	router.HandleFunc("/ping", WithLogging(dbPinger)).Methods("GET")
	router.HandleFunc("/ping", dbPinger).Methods("GET")

	router.Use(middles.GzipHandleEncoder)
	router.Use(middles.GzipHandleDecoder)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	return http.ListenAndServe(host, router)
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

// func LoadMS(memorial *MemStorage, fnam string) error {
// 	phil, err := os.OpenFile(fnam, os.O_RDONLY, 0666)
// 	//	phil, err := os.Open(fnam)
// 	if err != nil {
// 		return fmt.Errorf("file %s Open error %v", fnam, err)
// 	}
// 	defer phil.Close()
// 	reader := bufio.NewReader(phil)
// 	data, err := reader.ReadBytes('\n')
// 	if err != nil {
// 		return fmt.Errorf("file %s Read error %v", fnam, err)
// 	}
// 	err = memorial.UnmarshalMS(data)
// 	if err != nil {
// 		return fmt.Errorf(" Memstorage UnMarshal error %v", err)
// 	}
// 	return nil
// }

/*
metricstest -test.v -test.run="^TestIteration11[AB]*$" ^
-binary-path=cmd/server/server.exe -source-path=cmd/server/ ^
-agent-binary-path=cmd/agent/agent.exe ^
-server-port=8080 -file-storage-path=goshran.txt ^
-database-dsn=postgres://postgres:passwordas@localhost:5432/postgres


metricstest -test.v -test.run="^TestIteration1[AB]*$" -binary-path=cmd/server/server.exe -source-path=cmd/server/

go run . -d=postgres://postgres:passwordas@localhost:5432/postgres

*/
