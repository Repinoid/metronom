package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gorono/internal/memos"
	"os"
	"strconv"
)

// flags структура с параметрами агента в JSON файле
type flagAgent struct {
	Address        string `json:"address"`         // аналог переменной окружения ADDRESS или флага -a
	ReportInterval string `json:"report_interval"` // аналог переменной окружения STORE_INTERVAL или флага -i
	PollInterval   string `json:"poll_interval"`   // аналог переменной окружения STORE_INTERVAL или флага -i
	CryptoKey      string `json:"crypto_key"`      // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
	GrpcPort       string `json:"grpc_port"`       // gRPC, port
}

// initAgent() - инициализация параметров агента из аргументов командной строки
func initAgent() error {
	var err error

	var hostFlag, keyFlag, configFlag, grpcFlag string

	flag.StringVar(&configFlag, "c", "", "путь до файла с JSON конфигурации")
	flag.StringVar(&configFlag, "config", "", "путь до файла с JSON конфигурации") // -c = -config
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&keyFlag, "crypto-key", "", "путь до файла с публичным ключом")
	flag.StringVar(&grpcFlag, "g", gPort, "-g= GRPC port")
	reportIntervalFlag := flag.Int("r", reportInterval, "reportInterval")
	pollIntervalFlag := flag.Int("p", pollInterval, "pollIntervalFlag")
	rateLimitFlag := flag.Int("l", rateLimit, "pollIntervalFlag")
	flag.Parse()

	// с наименьшим приоритетом параметры агента из JSON файла
	if configFlag != "" {
		params, err := os.ReadFile(configFlag)
		if err != nil {
			return err
		}
		var prapor flagAgent
		err = json.Unmarshal(params, &prapor)
		if err != nil {
			return err
		}
		interval, err := memos.TakeInterval(prapor.PollInterval)
		if err != nil {
			return err
		}
		pollInterval = interval
		interval, err = memos.TakeInterval(prapor.ReportInterval)
		if err != nil {
			return err
		}
		reportInterval = interval

		host = prapor.Address
		cryptoKeyFile = prapor.CryptoKey
	}
	// параметры из флагов
	if hostFlag != "" {
		host = hostFlag
	}
	if keyFlag != "" {
		cryptoKeyFile = keyFlag
	}
	if grpcFlag != "" {
		gPort = grpcFlag
	}
	if *reportIntervalFlag != 0 {
		reportInterval = *reportIntervalFlag
	}
	if *pollIntervalFlag != 0 {
		pollInterval = *pollIntervalFlag
	}
	if *rateLimitFlag != 0 {
		rateLimit = *rateLimitFlag
	}
	// если есть переменные окружения - самый высокий приоритет
	enva, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = enva
	}
	enva, exists = os.LookupEnv("CRYPTO_KEY")
	if exists {
		cryptoKeyFile = enva
	}
	enva, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists {
		//		var err error
		reportInterval, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("REPORT_INTERVAL error value %s\t error %w", enva, err)
		}
	}
	enva, exists = os.LookupEnv("RATE_LIMIT")
	if exists {
		rateLimit, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("RATE_LIMIT error value %s\t error %w", enva, err)
		}
	}
	enva, exists = os.LookupEnv("POLL_INTERVAL")
	if exists {
		//		var err error
		pollInterval, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("POLL_INTERVAL error value %s\t error %w", enva, err)
		}
		return nil
	}

	if cryptoKeyFile != "" {
		// pkb - public key in []byte
		pkb, err := os.ReadFile(cryptoKeyFile)
		if err != nil {
			return err
		}
		cryptoKey = pkb
	} /* else {
		return fmt.Errorf("no public key file")
	}*/
	return nil
}
