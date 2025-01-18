package models

import (
	"context"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
type Gauge float64
type Counter int64

type Inter interface {
	GetMetric(ctx context.Context, metr *Metrics) (Metrics, error)
	PutMetric(ctx context.Context, metr *Metrics) error
	GetAllMetrics(ctx context.Context) (*[]Metrics, error)
	PutAllMetrics(ctx context.Context, metras *[]Metrics) error
	Ping(ctx context.Context) error
	LoadMS(fnam string) error
	SaveMS(fnam string) error
	Saver(fnam string, storeInterval int) error
}

func IsMetricsOK(metr Metrics) bool {
	if (metr.MType == "counter" && metr.Delta == nil) ||
		(metr.MType == "gauge" && metr.Value == nil) ||
		(metr.Delta != nil && metr.Value != nil) {
		return false
	}
	return true
}
