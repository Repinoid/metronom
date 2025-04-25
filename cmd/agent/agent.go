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
	cryptoKeyFile  = "cert.pem"
	publicKey      = ""
	rateLimit      = 4
	cunt           int64
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

	const chanCap = 4
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	//    signal.Notify(sigint, os.Interrupt)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigint
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	metroBarn := make(chan []models.Metrics, chanCap)
	go metrixIN(ctx, metroBarn)

	fenix := make(chan struct{})
	for w := 1; w <= rateLimit; w++ {
		go bolda(ctx, metroBarn, fenix)
	}
	// запуск нового работника если прежний вылетел по ошибке
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("cancel всему пришёл")
				// выход из бесконечного цикла найма работников
				return
			default:
				fenix <- struct{}{}             // блокируем канал пока балда не прочитает из него при своём завершении по ошибке
				go bolda(ctx, metroBarn, fenix) // нанимаем нового
			}
		}
	}()

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	// выполняем cancel контекста, metrixIN & bolda stop
	cancel()
	log.Println("Agent Shutdown gracefully")
	return nil
}

// получает банчи метрик и складывает в barn
func metrixIN(ctx context.Context, metroBarn chan<- []models.Metrics) {
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

// работник отсылает банчи метрик на сервер, феникс - канал для подачи сигнала о завершении по ошибке
func bolda(ctx context.Context, metroBarn <-chan []models.Metrics, fenix <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Горутина bolda остановлена")
			return

		default:
			bunch := <-metroBarn
			marshalledBunch, err := json.Marshal(bunch)
			// в случае ошибки читаем из феникса, разблокируя канал и выходим
			if err != nil {
				<-fenix
				return
			}
			//	var haHex string
			if publicKey != "" {
				//		keyB := md5.Sum([]byte(key))

				coded, err := privacy.Encrypt(marshalledBunch, []byte(publicKey))

				//		coded, err := privacy.EncryptB2B(marshalledBunch, keyB[:])
				if err != nil {
					<-fenix
					return
				}
				//		ha := privacy.MakeHash(nil, coded, keyB[:])
				//		haHex = hex.EncodeToString(ha)
				marshalledBunch = coded
			}
			compressedBunch, err := middlas.Pack2gzip(marshalledBunch)
			// в случае ошибки читаем из феникса, разблокируя канал и выходим

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

				// if key != "" {
			// 	req.Header.Add("HashSHA256", haHex)
			// }

			resp, _ := req.
				SetDoNotParseResponse(false).
				Post("/updates/") // slash on the tile
			if resp.StatusCode() == http.StatusOK { // при успешной отправке метрик обнуляем cчётчик
				atomic.StoreInt64(&cunt, 0) //	cunt = 0

			}

			log.Printf("AGENT responce from server %+v\n", resp.StatusCode())
		}
	}
}
