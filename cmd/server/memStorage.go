package main

import (
	"fmt"
	"strconv"
)

func (ms *MemStorage) addGauge(name string, value gauge) error {
	ms.mutter.Lock()
	defer ms.mutter.Unlock()
	ms.gau[name] = value
	return nil
}
func (ms *MemStorage) addCounter(name string, value counter) error {
	ms.mutter.Lock()
	defer ms.mutter.Unlock()
	if _, ok := ms.count[name]; ok {
		ms.count[name] += value
		return nil
	}
	ms.count[name] = value
	return nil
}
func (ms *MemStorage) getCounterValue(name string, value *string) error {
	ms.mutter.RLock() // <---- MUTEX
	defer ms.mutter.RUnlock()
	if _, ok := ms.count[name]; ok {
		*value = strconv.FormatInt(int64(ms.count[name]), 10)
		return nil
	}
	return fmt.Errorf("no %s key", name)
}
func (ms *MemStorage) getGaugeValue(name string, value *string) error {
	ms.mutter.RLock() // <---- MUTEX
	defer ms.mutter.RUnlock()
	if _, ok := ms.gau[name]; ok {
		*value = strconv.FormatFloat(float64(ms.gau[name]), 'f', -1, 64)
		return nil
	}
	return fmt.Errorf("no %s key", name)
}
