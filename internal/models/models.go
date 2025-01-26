// пакет определения типов
package models

import (
	"context"
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

type Inter interface {
	GetMetric(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	PutMetric(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	GetAllMetrics(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	PutAllMetrics(ctx context.Context, metr *Metrics, metras *[]Metrics) error
	Ping(ctx context.Context, dbepnt string) error
	LoadMS(fnam string) error
	SaveMS(fnam string) error
	Saver(fnam string, storeInterval int) error
	GetName() string
}

func IsMetricsOK(metr Metrics) bool {
	if (metr.MType != "gauge" && metr.MType != "counter") ||
		(metr.MType == "counter" && metr.Delta == nil) ||
		(metr.MType == "gauge" && metr.Value == nil) ||
		(metr.Delta != nil && metr.Value != nil) {
		return false
	}
	return true
}
