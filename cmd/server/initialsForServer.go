package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.uber.org/zap"

	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
)

// initServer () - инициализация параметров сервера и endpoint базы данных из аргументов командной строки
func InitServer() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	models.Logger = logger
	models.Sugar = *logger.Sugar()

	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		Host = hoster
		//		return nil
	}
	enva, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		var err error
		models.StoreInterval, err = strconv.Atoi(enva)
		if err != nil {
			log.Printf("STORE_INTERVAL error value %s\t error %v", enva, err)
		}
	}
	enva, exists = os.LookupEnv("KEY")
	if exists {
		models.Key = enva
	}
	enva, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists {
		models.FileStorePath = enva
	}
	enva, exists = os.LookupEnv("DATABASE_DSN")
	if exists {
		models.DBEndPoint = enva
	}
	enva, exists = os.LookupEnv("RESTORE")
	if exists {
		var err error
		models.ReStore, err = strconv.ParseBool(enva)
		if err != nil {
			log.Printf("RESTORE error value %s\t error %v", enva, err)
		}
		//	return nil
	}

	var hostFlag string
	var fileStoreFlag string
	var dbFlag string
	var keyFlag string

	flag.StringVar(&keyFlag, "k", models.Key, "KEY")
	flag.StringVar(&dbFlag, "d", models.DBEndPoint, "Data Base endpoint")
	flag.StringVar(&hostFlag, "a", Host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&fileStoreFlag, "f", models.FileStorePath, "-f= file to save memory storage")
	storeIntervalFlag := flag.Int("i", models.StoreInterval, "storeInterval")
	restoreFlag := flag.Bool("r", models.ReStore, "is restore mode on")

	flag.Parse()

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
	}
	if _, exists := os.LookupEnv("ADDRESS"); !exists {
		Host = hostFlag
	}
	if _, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		models.StoreInterval = *storeIntervalFlag
	}
	if _, exists := os.LookupEnv("FILE_STORAGE_PATH"); !exists {
		models.FileStorePath = fileStoreFlag
	}
	if _, exists := os.LookupEnv("RESTORE"); !exists {
		models.ReStore = *restoreFlag
	}
	if _, exists := os.LookupEnv("DATABASE_DSN"); !exists {
		models.DBEndPoint = dbFlag
	}
	if _, exists := os.LookupEnv("KEY"); !exists {
		models.Key = keyFlag
	}
	memStor := memos.InitMemoryStorage()

	if models.DBEndPoint == "" {
		log.Println("No base in Env variable and command line argument")
		models.Inter = memStor // если базы нет, подключаем in memory Storage
		return nil
	}

	ctx := context.Background()
	//	dbStorage, err := basis.InitDBStorage(ctx, "host=go_db user=postgres password=postgres dbname=postgres sslmode=disable")
	dbStorage, err := basis.InitDBStorage(ctx, models.DBEndPoint)

	if err != nil {
		models.Inter = memStor // если не удаётся подключиться к базе, подключаем in memory Storage
		log.Printf("Can't connect to DB %s\n", models.DBEndPoint)
		return nil
	}
	models.Inter = dbStorage // data base as Metric Storage
	return nil
}
