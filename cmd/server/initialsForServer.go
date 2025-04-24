package main

import (
	"context"
	"encoding/json"
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

// flags структура с параметрами сервера в JSON файле
type flagServer struct {
	Address        string `json:"address"`        // аналог переменной окружения ADDRESS или флага -a
	Restore        bool   `json:"restore"`        // аналог переменной окружения RESTORE или флага -r
	StoreInterval string `json:"store_interval"` // аналог переменной окружения STORE_INTERVAL или флага -i
	StoreFile     string `json:"store_file"`     // аналог переменной окружения STORE_FILE или -f
	DatabaseDsn   string `json:"database_dsn"`   // аналог переменной окружения DATABASE_DSN или флага -d
	CryptoKey     string `json:"crypto_key"`     // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}

// initServer () - инициализация параметров сервера и endpoint базы данных из аргументов командной строки
func InitServer() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	models.Logger = logger
	models.Sugar = *logger.Sugar()

	// возрастание приоритетов - из JSON файла, флаги, переменные окружения
	// будем инициализировать параметры в этом порядке, более приоритетные переопределят предыдущие присваивания
	var hostFlag string
	var fileStoreFlag string
	var dbFlag string
	var keyFlag string
	var configFlag string

	flag.StringVar(&configFlag, "c", "", "путь до файла с JSON конфигурации")	// "" - по умолчанию пусто
	flag.StringVar(&configFlag, "config", "", "путь до файла с JSON конфигурации") // -c = -config
	flag.StringVar(&keyFlag, "crypto-key", models.Key, "путь до файла с private ключом")
	flag.StringVar(&dbFlag, "d", models.DBEndPoint, "Data Base endpoint")
	flag.StringVar(&hostFlag, "a", Host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&fileStoreFlag, "f", models.FileStorePath, "-f= file to save memory storage")
	storeIntervalFlag := flag.Int("i", models.StoreInterval, "storeInterval")
	restoreFlag := flag.Bool("r", models.ReStore, "is restore mode on")

	flag.Parse()

	// если есть JSON файл с параметрами
	if configFlag != "" {
		params, err := os.ReadFile(configFlag)
		if err != nil {
			return err
		}
		var prapor flagServer
		err = json.Unmarshal(params, &prapor)
		if err != nil {
			return err
		}
		models.ReStore = prapor.Restore // bool, надо ли сохранять в файл
		if prapor.Restore {
			interval, err := memos.TakeInterval(prapor.StoreInterval)
			if err != nil {
				return err
			}
			models.StoreInterval = interval
		}
		Host = prapor.Address
		models.Key = prapor.CryptoKey
		models.FileStorePath = prapor.StoreFile
		models.DBEndPoint = prapor.DatabaseDsn
	}
	// параметры из флагов командной строки, которые есть
	if hostFlag != "" {
		Host = hostFlag
	}
	if *storeIntervalFlag != 0 {
		models.StoreInterval = *storeIntervalFlag
	}
	if fileStoreFlag != "" {
		models.FileStorePath = fileStoreFlag
	}
	if *restoreFlag {
		models.ReStore = *restoreFlag
	}
	if dbFlag != "" {
		models.DBEndPoint = dbFlag
	}
	if keyFlag != "" {
		models.Key = keyFlag
	}
	// параметры из переменных окружения
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		Host = hoster
	}
	enva, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		var err error
		models.StoreInterval, err = strconv.Atoi(enva)
		if err != nil {
			log.Printf("STORE_INTERVAL error value %s\t error %v", enva, err)
		}
	}
	enva, exists = os.LookupEnv("CRYPTO_KEY")
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
	}

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
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

	if models.Key != "" {
		// pkb - private key in []byte
		pkb, err := os.ReadFile(models.Key)
		if err != nil {
			return err
		}
		models.PrivateKey = string(pkb)
	}

	return nil
}
