package main

import (
	"context"
	"flag"
	"fmt"
	"internal/dbaser"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
)

var storeInterval = 300
var fileStorePath = "./goshran.txt"
var reStore = true
var dbEndPoint = ""

func foa4Server() error {
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
	}
	enva, exists := os.LookupEnv("STORE_INTERVAL")
	if exists {
		var err error
		storeInterval, err = strconv.Atoi(enva)
		if err != nil {
			log.Printf("STORE_INTERVAL error value %s\t error %v", enva, err)
		}
	}
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

	flag.StringVar(&dbFlag, "d", dbEndPoint, "Data Base endpoint")
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&fileStoreFlag, "f", fileStorePath, "Only -a={host:port} flag is allowed here")
	storeIntervalFlag := flag.Int("i", storeInterval, "storeInterval")
	restoreFlag := flag.Bool("r", reStore, "restore")

	flag.Parse()

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
	}
	// if no environment variables - load from arguments
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
	if dbEndPoint == "" {
		log.Println("No base in Env variable and command line argument")
		MetricBaseStruct.IsBase = false
		return nil
	}
	ctx := context.Background()
	mb, err := pgx.Connect(ctx, dbEndPoint)
	//	mb.PgConn().
	MetricBaseStruct = dbaser.Struct4db{MetricBase: mb, Ctx: ctx, IsBase: false}
	if err != nil {
		log.Printf("Can't connect to DB %s\n", dbEndPoint)
		return nil
	}
	err = dbaser.TableCreation(&MetricBaseStruct)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create tables: %v\n", err)
		return nil
	}
	MetricBaseStruct.IsBase = true

	return nil
}
