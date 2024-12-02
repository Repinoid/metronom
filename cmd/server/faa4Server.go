package main

import (
	"flag"
	"os"
)

func faa4server() int {
	hoster, exists := os.LookupEnv("ADDRESS")
	if exists {
		host = hoster
		return 0
	}
	var hostFlag string
	flag.StringVar(&hostFlag, "a", "localhost:8080", "Only -a={host:port} flag is allowed here")
	flag.Parse()

	if hostFlag == "" {
		return 1
	}

	host = hostFlag

	return 0
}
