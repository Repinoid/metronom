package main

import (
	"flag"
	"fmt"
	"os"
)

func foa4Server() error {
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
		return nil
	}
	var hostFlag string
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.Parse()

	if hostFlag == "" {
		return fmt.Errorf("no host parsed from arg string")
	}

	host = hostFlag

	return nil
}
