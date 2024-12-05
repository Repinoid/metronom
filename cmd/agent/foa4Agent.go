package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

func foa4Agent() error {
	enva, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = enva
	}
	enva, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists {
		var err error
		reportInterval, err  = strconv.Atoi(enva)
		if err != nil {
			log.Fatalf("REPORT_INTERVAL error value %s\t error %v\n", enva, err)
		}
	}
	enva, exists = os.LookupEnv("POLL_INTERVAL")
	if exists {
		var err error
		pollInterval, err = strconv.Atoi(enva)
		if err != nil {
			log.Fatalf("POLL_INTERVAL error value %s\t error %v\n", enva, err)
		}
	}

	var hostFlag string
	flag.StringVar(&hostFlag, "a", "localhost:8080", "Only -a={host:port} flag is allowed here")
	reportIntervalFlag := flag.Int("r", reportInterval, "reportInterval")
	pollIntervalFlag := flag.Int("p", pollInterval, "pollIntervalFlag")
	flag.Parse()

	if _, exists := os.LookupEnv("ADDRESS"); !exists {
		host = hostFlag
	}
	if _, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		reportInterval = *reportIntervalFlag
	}
	if _, exists := os.LookupEnv("POLL_INTERVAL"); !exists {
		pollInterval = *pollIntervalFlag
	}
	return nil
}
