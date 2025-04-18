package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func initLoader() error {
	enva, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = enva
	}
	enva, exists = os.LookupEnv("KEY")
	if exists {
		key = enva
	}
	enva, exists = os.LookupEnv("REPORT_INTERVAL")
	if exists {
		var err error
		reportInterval, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("REPORT_INTERVAL error value %s\t error %w", enva, err)
		}
	}
	enva, exists = os.LookupEnv("RATE_LIMIT")
	if exists {
		var err error
		rateLimit, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("RATE_LIMIT error value %s\t error %w", enva, err)
		}
	}

	var hostFlag, keyFlag string
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&keyFlag, "k", key, "int")
	reportIntervalFlag := flag.Int("r", reportInterval, "reportInterval")
	rateLimitFlag := flag.Int("l", rateLimit, "pollIntervalFlag")
	flag.Parse()

	if _, exists := os.LookupEnv("ADDRESS"); !exists {
		host = hostFlag
	}
	if _, exists := os.LookupEnv("KEY"); !exists {
		key = keyFlag
	}
	if _, exists := os.LookupEnv("REPORT_INTERVAL"); !exists {
		reportInterval = *reportIntervalFlag
	}
	if _, exists := os.LookupEnv("RATE_LIMIT"); !exists {
		rateLimit = *rateLimitFlag
	}
	return nil
}
