package main

import (
	"net/http"
	"strconv"
)

func newMemStorage() MemStorage {
	memStor := new(MemStorage)
	memStor.gau = make(map[string]gauge)
	memStor.count = make(map[string]counter)
	return *memStor
}
func (ms *MemStorage) addGauge(name string, value gauge) error {
	ms.gau[name] = value
	return nil
}
func (ms *MemStorage) addCounter(name string, value counter) error {
	if _, ok := ms.count[name]; ok {
		ms.count[name] += value
		return nil
	}
	ms.count[name] = value
	return nil
}
func (ms *MemStorage) getCounterValue(name string, value *string) int {
	if _, ok := ms.count[name]; ok {
		*value = strconv.FormatInt(int64(ms.count[name]), 10)
		return http.StatusOK
	}
	return http.StatusNotFound
}
func (ms *MemStorage) getGaugeValue(name string, value *string) int {
	if _, ok := ms.gau[name]; ok {
		*value = strconv.FormatFloat(float64(ms.gau[name]), 'f', -1, 64)
		return http.StatusOK
	}
	return http.StatusNotFound
}
