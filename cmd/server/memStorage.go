package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

type MStorJSON struct {
	Gau   map[string]gauge
	Count map[string]counter
}

func (memorial *MemStorage) UnmarshalMS(data []byte) error {
	memor := MStorJSON{}
	buf := bytes.NewBuffer(data)
	memorial.mutter.Lock()
	err := json.NewDecoder(buf).Decode(&memor)
	memorial.gau = memor.Gau
	memorial.count = memor.Count
	memorial.mutter.Unlock()
	return err
}
func (memorial *MemStorage) MarshalMS() ([]byte, error) {
	buf := new(bytes.Buffer)
	memorial.mutter.RLock()
	err := json.NewEncoder(buf).Encode(MStorJSON{
		Gau:   memorial.gau,
		Count: memorial.count,
	})
	memorial.mutter.RUnlock()
	return append(buf.Bytes(), '\n'), err
}

func (memorial *MemStorage) LoadMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	reader := bufio.NewReader(phil)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("file %s Read error %v", fnam, err)
	}
	err = memorial.UnmarshalMS(data)
	if err != nil {
		return fmt.Errorf(" Memstorage UnMarshal error %v", err)
	}
	return nil
}
func (memorial *MemStorage) SaveMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	march, err := memorial.MarshalMS()
	if err != nil {
		return fmt.Errorf(" Memstorage Marshal error %v", err)
	}
	_, err = phil.Write(march)
	if err != nil {
		return fmt.Errorf("file %s Write error %v", fnam, err)
	}
	return nil
}
