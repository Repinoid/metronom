package main

import (
	"fmt"
)

func (memorial *MemStorage) addGauge(name string, value gauge) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()
	memorial.gau[name] = value
	return nil
}
func (memorial *MemStorage) addCounter(name string, value counter) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()
	if _, ok := memorial.count[name]; ok {
		memorial.count[name] += value
		return nil
	}
	memorial.count[name] = value
	return nil
}
func (memorial *MemStorage) getCounterValue(name string, value *counter) error {
	memorial.mutter.RLock() // <---- MUTEX
	defer memorial.mutter.RUnlock()
	if _, ok := memorial.count[name]; ok {
		*value = memorial.count[name]
		return nil
	}
	return fmt.Errorf("no %s key", name)
}
func (memorial *MemStorage) getGaugeValue(name string, value *gauge) error {
	memorial.mutter.RLock() // <---- MUTEX
	defer memorial.mutter.RUnlock()
	if _, ok := memorial.gau[name]; ok {
		*value = memorial.gau[name]
		return nil
	}
	return fmt.Errorf("no %s key", name)
}
