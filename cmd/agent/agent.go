// агент (HTTP-клиент) для сбора рантайм-метрик и их последующей отправки на сервер по протоколу HTTP.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"
)

var host = "localhost:8080"

var (
	reportInterval = 10
	pollInterval   = 2
	//	cryptoKeyFile  = ""
	cryptoKeyFile = "cert.pem"
	cryptoKey     = []byte("")
	rateLimit     = 3
	cunt          int64
)

// Глобальные переменные для флага компилляции.
// Форма запуска go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X main.buildCommit=comitta" main.go
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := initAgent(); err != nil {
		log.Fatal(err)
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

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-exit
		cancel()
	}()

	var wg sync.WaitGroup

	const chanCap = 4

	metroBarn := make(chan []models.Metrics, chanCap)

	wg.Add(1)
	go metrixIN(ctx, metroBarn, &wg)

	for w := 1; w <= rateLimit; w++ {
		wg.Add(1)
		log.Println("Балда запущена")
		go bolda(ctx, metroBarn, &wg)
	}

	wg.Wait()
	close(metroBarn)
	log.Println("Agent Shutdown gracefully")
	return nil
}

// получает банчи метрик и складывает в barn
// func metrixIN(ctx context.Context, metroBarn chan<- []models.Metrics, wg *sync.WaitGroup, sigint chan os.Signal) {
func metrixIN(ctx context.Context, metroBarn chan<- []models.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	memStorage := []models.Metrics{}
	tickerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(reportInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("Горутина metrixIN остановлена")
			return
		// по тикеру запрашиваем метрикис рантайма
		case <-tickerPoll.C:
			memStorage = *memos.GetMetrixFromOS()
			addMetrix := *memos.GetMoreMetrix()
			memStorage = append(memStorage, addMetrix...)
			atomic.AddInt64(&cunt, 1) //			cunt++

			// search for PollCount metric
			for ind, metr := range memStorage {
				if metr.ID == "PollCount" && metr.MType == "counter" {
					cu := atomic.LoadInt64(&cunt)
					memStorage[ind].Delta = &cu // memStorage[ind].Delta = cunt
					break
				}
			}
		// засылаем метрики в канал
		case <-tickerReport.C:
			metroBarn <- memStorage
		}
	}
}

// работник отсылает банчи метрик на сервер,
func bolda(ctx context.Context, metroBarn <-chan []models.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	var bunch []models.Metrics
	for {
		select {
		case <-ctx.Done():
			log.Println("Горутина bolda остановлена")

			// search for PollCount metric, if not 0 - sendMetrics напоследок
			for ind, metr := range bunch {
				if metr.ID == "PollCount" && metr.MType == "counter" {
					if *bunch[ind].Delta > 0 {
						sendMetrics(bunch)
					}
					break // нашлось
				}
			}
			return
		case bunch = <-metroBarn:
			err := sendMetrics(bunch)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// sendMetrics посылает слайс метрик на сервер
func sendMetrics(bunch []models.Metrics) (err error) {

	marshalledBunch, err := json.Marshal(bunch)
	if err != nil {
		log.Println(err)
		return
	}
	if cryptoKeyFile != "" {

		coded, err := privacy.Encrypt(marshalledBunch, cryptoKey)

		if err != nil {
			log.Println(err)
			return err
		}
		marshalledBunch = coded
	}
	compressedBunch, err := middlas.Pack2gzip(marshalledBunch)

	if err != nil {
		log.Println(err)
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

	resp, _ := req.
		SetDoNotParseResponse(false).
		Post("/updates/") // slash on the tile
	if resp.StatusCode() == http.StatusOK { // при успешной отправке метрик обнуляем cчётчик
		atomic.StoreInt64(&cunt, 0) //	cunt = 0

	}

	log.Printf("AGENT responce from server %+v\n", resp.StatusCode())

	return nil
}
