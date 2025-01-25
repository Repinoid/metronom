package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"

	"github.com/go-resty/resty/v2"
)

var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2
var key = ""
var rateLimit = 4

func main() {
	if err := initAgent(); err != nil {
		log.Fatal("INTERVALS error ", err)
		return
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	const chanCap = 4
	const rabNum = 4

	metroBarn := make(chan []models.Metrics, chanCap)
	go metrixIN(metroBarn)

	for w := 1; w <= rabNum; w++ {
		go bolda(metroBarn)
	}
	stopper := make(chan<- struct{})
	stopper <- struct{}{}
	return nil
}

// получает банчи метрик и складывает в barn
func metrixIN(metroBarn chan<- []models.Metrics) {
	for {
		memStorage := []models.Metrics{}
		cunt := int64(0)
		for i := 0; i < reportInterval/pollInterval; i++ {
			memStorage = *memos.GetMetrixFromOS()
			addMetrix := *memos.GetMoreMetrix()
			memStorage = append(memStorage, addMetrix...)
			cunt++
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
		for ind, metr := range memStorage {
			if metr.ID == "PollCount" && metr.MType == "counter" {
				memStorage[ind].Delta = &cunt // в сам memStorage, metr - копия
				break
			}
		}
		metroBarn <- memStorage
	}
}

// работник отсылает банчи метрик на сервер
func bolda(metroBarn <-chan []models.Metrics) {
	for {
		bunch := <-metroBarn
		marshalledBunch, err := json.Marshal(bunch)
		if err != nil {
			return
		}
		var haHex string
		if key != "" {
			keyB := md5.Sum([]byte(key))

			coded, err := privacy.EncryptB2B(marshalledBunch, keyB[:])
			if err != nil {
				return
			}
			ha := privacy.MakeHash(nil, coded, keyB[:])
			haHex = hex.EncodeToString(ha)
			marshalledBunch = coded
		}
		compressedBunch, err := middlas.Pack2gzip(marshalledBunch)
		if err != nil {
			return
		}

		httpc := resty.New() //
		httpc.SetBaseURL("http://" + host)

		httpc.SetRetryCount(3)
		httpc.SetRetryWaitTime(1 * time.Second)    // начальное время повтора
		httpc.SetRetryMaxWaitTime(9 * time.Second) // 1+3+5
		httpc.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			rwt := client.RetryWaitTime
			client.SetRetryWaitTime(rwt + 2*time.Second) //	увеличение времени ожидания на 2 сек
			return client.RetryWaitTime, nil
		})

		req := httpc.R().
			SetHeader("Content-Encoding", "gzip"). // сжаtо
			SetBody(compressedBunch).
			SetHeader("Accept-Encoding", "gzip")

		if key != "" {
			req.Header.Add("HashSHA256", haHex)
		}

		resp, _ := req.
			SetDoNotParseResponse(false).
			Post("/updates/") // slash on the tile

		log.Printf("AGENT responce from server %+v\n", resp.StatusCode())
	}
}
