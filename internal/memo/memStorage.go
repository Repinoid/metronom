package memo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"internal/dbaser"
	"log"
	"sync"

	"os"
)

type Gauge float64
type Counter int64
type gauge = Gauge
type counter = Counter
type MemStorage struct {
	Gaugemetr map[string]gauge
	Countmetr map[string]counter
	Mutter    sync.RWMutex
}

// func (memorial *MemStorage) AddGauge(name string, value gauge) error {
func AddGauge(memorial *MemStorage, baza dbaser.Struct4db, name string, value gauge) error {
	if baza.IsBase {
		err := dbaser.TablePutGauge(baza.Ctx, baza.MetricBase, name, float64(value))
		if err != nil {
			log.Printf("from memstorage %v\nisBase - %v\\n\n\n", memorial, baza)
			//sugar.Errorf("err", err)
		}
	}
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	memorial.Gaugemetr[name] = value
	return nil
}
func AddCounter(memorial *MemStorage, baza dbaser.Struct4db, name string, value counter) error {
	if baza.IsBase {
		err := dbaser.TablePutCounter(baza.Ctx, baza.MetricBase, name, int64(value))
		if err != nil {
			log.Printf("from memstorage %v\nisBase - %v\\n\n\n", memorial, baza)
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
	//func (memorial *MemStorage) GetCounterValue(name string, value *counter) error {
	if baza.IsBase {
		cunt, err := dbaser.TableGetCounter(baza.Ctx, baza.MetricBase, name)
		if err == nil {
			*value = counter(cunt)
			return nil
		}
		log.Printf("from memstorage %v\nisBase - %v\\n\n\n", memorial, baza)
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
	//func (memorial *MemStorage) GetGaugeValue(name string, value *gauge) error {
	if baza.IsBase {
		gaaga, err := dbaser.TableGetGauge(baza.Ctx, baza.MetricBase, name)
		if err == nil {
			*value = gauge(gaaga)
			return nil
		}
		log.Printf("from memstorage %v\nisBase - %v\\n\n\n", memorial, baza)
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
