package main

import (
	"context"
	"flag"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/memos"
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"
)

var storeInterval = 300
var fileStorePath = "./goshran.txt"
var reStore = true
var dbEndPoint = ""
var key string = ""

func InitServer() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
		//		return nil
	}
	enva, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		var err error
		storeInterval, err = strconv.Atoi(enva)
		if err != nil {
			log.Printf("STORE_INTERVAL error value %s\t error %v", enva, err)
		}
	}
	// enva, exists = os.LookupEnv("KEY")
	// if exists {
	// 	key = enva
	// }
	enva, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists {
		fileStorePath = enva
	}
	enva, exists = os.LookupEnv("DATABASE_DSN")
	if exists {
		dbEndPoint = enva
	}
	enva, exists = os.LookupEnv("RESTORE")
	if exists {
		var err error
		reStore, err = strconv.ParseBool(enva)
		if err != nil {
			log.Printf("RESTORE error value %s\t error %v", enva, err)
		}
		//	return nil
	}

	var hostFlag string
	var fileStoreFlag string
	var dbFlag string
	var keyFlag string

	flag.StringVar(&keyFlag, "k", key, "KEY")
	flag.StringVar(&dbFlag, "d", dbEndPoint, "Data Base endpoint")
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&fileStoreFlag, "f", fileStorePath, "Only -a={host:port} flag is allowed here")
	storeIntervalFlag := flag.Int("i", storeInterval, "storeInterval")
	restoreFlag := flag.Bool("r", reStore, "restore")

	flag.Parse()

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
	}
	if _, exists := os.LookupEnv("ADDRESS"); !exists {
		host = hostFlag
	}
	if _, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		storeInterval = *storeIntervalFlag
	}
	if _, exists := os.LookupEnv("FILE_STORAGE_PATH"); !exists {
		fileStorePath = fileStoreFlag
	}
	if _, exists := os.LookupEnv("RESTORE"); !exists {
		reStore = *restoreFlag
	}
	if _, exists := os.LookupEnv("DATABASE_DSN"); !exists {
		dbEndPoint = dbFlag
	}
	if _, exists := os.LookupEnv("KEY"); !exists {
		key = keyFlag
	}
	memStor := memos.InitMemoryStorage()

	if dbEndPoint == "" {
		log.Println("No base in Env variable and command line argument")
		inter = memStor // если базы нет, подключаем in memory Storage
		return nil
	}

	ctx = context.Background()
	dbStorage, err := basis.InitDBStorage(ctx, dbEndPoint)

	if err != nil {
		inter = memStor // если не удаётся подключиться к базе, подключаем in memory Storage
		log.Printf("Can't connect to DB %s\n", dbEndPoint)
		return nil
	}
	inter = dbStorage // data base as Metric Storage
	return nil
}
