// пакет определения типов
package models

import (
	"context"

	"go.uber.org/zap"
)

// struct - имя, тип, значение (ссылка на)
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type Gauge float64
type Counter int64

var Inter Interferon

type Interferon interface {
	GetMetric(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	PutMetric(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	GetAllMetrics(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	PutAllMetrics(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	Ping(ctx context.Context, dbepnt string) error
	LoadMS(fnam string) error
	SaveMS(fnam string) error
	Saver(ctx context.Context, fnam string, storeInterval int) error
	GetName() string
	Close()
}

// переменные для логгера и получающие значения из аргументов командной строки
var (
	Sugar  zap.SugaredLogger
	Logger *zap.Logger
	// параметры из флагов командной строки
	StoreInterval = 300             // интервал сохранения метрик в файл, секунд
	FileStorePath = "./goshran.txt" // файл для сохранения метрик
	ReStore       = true            // надо ли сохранять метрики в файл
	DBEndPoint    = ""              // endpoint базы данных
	//	Key           string = ""              // file name with private key
	Key        string = "privateKey.pem" // file name with private key
	PrivateKey string = ""               // private key, содержится в файле с именем в Key

)
