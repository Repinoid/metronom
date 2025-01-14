package memo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"internal/dbaser"
	"sync"

	"os"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Gauge float64
type Counter int64
type gauge = Gauge
type counter = Counter
type MemStorage struct {
	Gaugemetr map[string]gauge
	Countmetr map[string]counter
	Mutter    sync.RWMutex
}

func AddGauge(memorial *MemStorage, baza dbaser.Struct4db, name string, value gauge) error {
	if baza.IsBase {
		valadr := float64(value)
		metro := dbaser.Metrics{ID: name, MType: "counter", Value: &valadr}
		err := dbaser.TableMetricWrapper(dbaser.TablePutMetric)(&baza, &metro)
		if err != nil {
			return fmt.Errorf("AddGauge err name %s value %g baza %+v err %w\n", name, value, baza, err)
		}
	}
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	memorial.Gaugemetr[name] = value
	return nil
}
func AddCounter(memorial *MemStorage, baza dbaser.Struct4db, name string, value counter) error {
	if baza.IsBase {
		valadr := int64(value)
		metro := dbaser.Metrics{ID: name, MType: "counter", Delta: &valadr}
		err := dbaser.TableMetricWrapper(dbaser.TablePutMetric)(&baza, &metro)
		if err != nil {
			return fmt.Errorf("AddCounter err name %s value %d baza %+v err %w\n", name, value, baza, err)
		}
	}
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	if _, ok := memorial.Countmetr[name]; ok {
		memorial.Countmetr[name] += value
		return nil
	}
	memorial.Countmetr[name] = value
	return nil
}
func GetCounterValue(memorial *MemStorage, baza dbaser.Struct4db, name string, value *counter) error {
	if baza.IsBase {
		var cunt int64
		metro := dbaser.Metrics{ID: name, MType: "counter", Delta: &cunt}
		//	err := dbaser.TableGetCounter(&baza, name, &cunt)
		err := dbaser.TableMetricWrapper(dbaser.TableGetMetric)(&baza, &metro)
		if err == nil {
			*value = counter(cunt)
			return nil
		}
		return fmt.Errorf("GetCounter err name %s value %d baza %+v err %w\n", name, value, baza, err)
	}
	memorial.Mutter.RLock() // <---- MUTEX
	defer memorial.Mutter.RUnlock()
	if _, ok := memorial.Countmetr[name]; ok {
		*value = memorial.Countmetr[name]
		return nil
	}
	return fmt.Errorf("no %s key", name)
}
func GetGaugeValue(memorial *MemStorage, baza dbaser.Struct4db, name string, value *gauge) error {
	if baza.IsBase {
		var gaaga float64
		//		var gaaga float64
		metro := dbaser.Metrics{ID: name, MType: "gauge", Value: &gaaga}
		//		err := dbaser.TableGetGauge(&baza, name, &gaaga)
		err := dbaser.TableMetricWrapper(dbaser.TableGetMetric)(&baza, &metro)
		if err == nil {
			*value = gauge(*metro.Value)
			//			*value = gauge(gaaga)
			return nil
		}
		return fmt.Errorf("GetGauge err name %s value %d baza %+v err %w\n", name, value, baza, err)
	}
	memorial.Mutter.RLock() // <---- MUTEX
	defer memorial.Mutter.RUnlock()
	if _, ok := memorial.Gaugemetr[name]; ok {
		*value = memorial.Gaugemetr[name]
		return nil
	}
	return fmt.Errorf("no %s key", name)
}

type MStorJSON struct {
	Gaugemetr map[string]gauge
	Countmetr map[string]counter
}

func (memorial *MemStorage) UnmarshalMS(data []byte) error {
	memor := MStorJSON{
		Gaugemetr: make(map[string]gauge),
		Countmetr: make(map[string]counter),
	}
	buf := bytes.NewBuffer(data)
	memorial.Mutter.Lock()
	err := json.NewDecoder(buf).Decode(&memor)
	memorial.Gaugemetr = memor.Gaugemetr
	memorial.Countmetr = memor.Countmetr
	memorial.Mutter.Unlock()
	return err
}
func (memorial *MemStorage) MarshalMS() ([]byte, error) {
	buf := new(bytes.Buffer)
	memorial.Mutter.RLock()
	err := json.NewEncoder(buf).Encode(MStorJSON{
		Gaugemetr: memorial.Gaugemetr,
		Countmetr: memorial.Countmetr,
	})
	memorial.Mutter.RUnlock()
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
