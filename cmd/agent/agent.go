// агент (HTTP-клиент) для сбора рантайм-метрик и их последующей отправки на сервер по протоколу HTTP.
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"
)

//var host = ":8080"

var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2
var key = ""
var rateLimit = 4
var cunt int64

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := initAgent(); err != nil {
		log.Fatal("INTERVALS error ", err)
		return
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	const chanCap = 4

	metroBarn := make(chan []models.Metrics, chanCap)
	go metrixIN(metroBarn)

	fenix := make(chan struct{})
	for w := 1; w <= rateLimit; w++ {
		go bolda(metroBarn, fenix)
	}
	for {
		fenix <- struct{}{}        // блокируем канал пока балда не прочитает из него при своём завершении по ошибке
		go bolda(metroBarn, fenix) // нанимаем нового
	}
}

// получает банчи метрик и складывает в barn
func metrixIN(metroBarn chan<- []models.Metrics) {
	memStorage := []models.Metrics{}
	tickerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for {
		select {
		case <-tickerPoll.C:
			memStorage = *memos.GetMetrixFromOS()
			addMetrix := *memos.GetMoreMetrix()
			memStorage = append(memStorage, addMetrix...)
			atomic.AddInt64(&cunt, 1) //			cunt++

			for ind, metr := range memStorage {
				if metr.ID == "PollCount" && metr.MType == "counter" { // search for PollCount metric
					cu := atomic.LoadInt64(&cunt)
					memStorage[ind].Delta = &cu // memStorage[ind].Delta = cunt
					break
				}
			}
		case <-tickerReport.C:
			metroBarn <- memStorage
		}
	}
}

// работник отсылает банчи метрик на сервер, феникс - канал для подачи сигнала о завершении по ошибке
func bolda(metroBarn <-chan []models.Metrics, fenix <-chan struct{}) {
	for {
		bunch := <-metroBarn
		marshalledBunch, err := json.Marshal(bunch)
		if err != nil {
			<-fenix // в случае ошибки читаем из феникса, разблокируя канал и выходим
			return
		}
		var haHex string
		if key != "" {
			keyB := md5.Sum([]byte(key))

			coded, err := privacy.EncryptB2B(marshalledBunch, keyB[:])
			if err != nil {
				<-fenix
				return
			}
			ha := privacy.MakeHash(nil, coded, keyB[:])
			haHex = hex.EncodeToString(ha)
			marshalledBunch = coded
		}
		compressedBunch, err := middlas.Pack2gzip(marshalledBunch)
		if err != nil {
			<-fenix
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
		if resp.StatusCode() == http.StatusOK { // при успешной отправке метрик обнуляем cчётчик
			atomic.StoreInt64(&cunt, 0) //	cunt = 0

		}

		log.Printf("AGENT responce from server %+v\n", resp.StatusCode())
	}
}
