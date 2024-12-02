package main

import (
	"flag"
	"os"
	"strconv"
)

func faa4Agent() int {
	enva, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = enva
	}
	enva, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists {
		reportInterval, _ = strconv.Atoi(enva)
	}
	enva, exists = os.LookupEnv("POLL_INTERVAL")
	if exists {
		pollInterval, _ = strconv.Atoi(enva)
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
	return 0
}
