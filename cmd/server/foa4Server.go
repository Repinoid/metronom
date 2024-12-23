package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var STORE_INTERVAL = 300
var FILE_STORAGE_PATH = "./goshran.txt"
var RESTORE = true

func foa4Server() error {
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
		return nil
	}
	enva, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		var err error
		STORE_INTERVAL, err = strconv.Atoi(enva)
		if err != nil {
			return fmt.Errorf("STORE_INTERVAL error value %s\t error %w", enva, err)
		}
	}
	enva, exists = os.LookupEnv("FILE_STORAGE_PATH")
	if exists {
		var err error
		FILE_STORAGE_PATH = enva
		if err != nil {
			return fmt.Errorf("FILE_STORAGE_PATH error value %s\t error %w", enva, err)
		}
		return nil
	}
	enva, exists = os.LookupEnv("RESTORE")
	if exists {
		var err error
		RESTORE, err = strconv.ParseBool(enva)
		if err != nil {
			return fmt.Errorf("RESTORE error value %s\t error %w", enva, err)
		}
		return nil
	}

	var hostFlag string
	var fileStoreFlag string

	flag.StringVar(&hostFlag, "a", "localhost:8080", "Only -a={host:port} flag is allowed here")
	flag.StringVar(&fileStoreFlag, "f", FILE_STORAGE_PATH, "Only -a={host:port} flag is allowed here")
	storeIntervalFlag := flag.Int("i", STORE_INTERVAL, "storeInterval")
	restoreFlag := flag.Bool("r", RESTORE, "restore")

	flag.Parse()

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
	}
	if _, exists := os.LookupEnv("ADDRESS"); !exists {
		host = hostFlag
	}
	if _, exists := os.LookupEnv("STORE_INTERVAL"); !exists {
		STORE_INTERVAL = *storeIntervalFlag
	}
	if _, exists := os.LookupEnv("FILE_STORAGE_PATH"); !exists {
		FILE_STORAGE_PATH = fileStoreFlag
	}
	if _, exists := os.LookupEnv("RESTORE"); !exists {
		RESTORE = *restoreFlag
	}
	return nil
}
