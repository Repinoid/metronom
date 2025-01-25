package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func initAgent() error {
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
	enva, exists = os.LookupEnv("POLL_INTERVAL")
	if exists {
		var err error
		pollInterval, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("POLL_INTERVAL error value %s\t error %w", enva, err)
		}
		return nil
	}

	var hostFlag string
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&key, "k", key, "Only -a={host:port} flag is allowed here")
	reportIntervalFlag := flag.Int("r", reportInterval, "reportInterval")
	pollIntervalFlag := flag.Int("p", pollInterval, "pollIntervalFlag")
	rateLimitFlag := flag.Int("l", pollInterval, "pollIntervalFlag")
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
	if _, exists := os.LookupEnv("RATE_LIMIT"); !exists {
		rateLimit = *rateLimitFlag
	}
	return nil
}
