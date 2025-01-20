package memos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gorono/internal/models"
	"log"
	"os"
	"sync"
	"time"
)

type MemoryStorageStruct struct {
	Gaugemetr map[string]models.Gauge
	Countmetr map[string]models.Counter
	Mutter    *sync.RWMutex
}
type Metrics = models.Metrics
var mtx sync.RWMutex

func InitMemoryStorage() *MemoryStorageStruct {
	memStor := MemoryStorageStruct{
		Gaugemetr: make(map[string]gauge),
		Countmetr: make(map[string]counter),
		Mutter:    &mtx,
	}
	return &memStor
}

func (memorial *MemoryStorageStruct) PutMetric(ctx context.Context, metr *Metrics) error {
	if !models.IsMetricsOK(*metr) {
		return fmt.Errorf("bad metric %+v", metr)
	}
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	switch metr.MType {
	case "gauge":
		memorial.Gaugemetr[metr.ID] = models.Gauge(*metr.Value)
	case "counter":
		memorial.Countmetr[metr.ID] += models.Counter(*metr.Delta)
	default:
		return fmt.Errorf("wrong type %s", metr.MType)
	}
	return nil
}

func (memorial *MemoryStorageStruct) GetMetric(ctx context.Context, metr *Metrics) (Metrics, error) {
	memorial.Mutter.RLock() // <---- MUTEX
	defer memorial.Mutter.RUnlock()
	metrix := Metrics{ID: metr.ID, MType: metr.MType} // new pure Metrics to return, nil Delta&Value ptrs
	switch metr.MType {
	case "gauge":
		if val, ok := memorial.Gaugemetr[metr.ID]; ok {
			out := float64(val)
			metrix.Value = &out
			break
		}
		return metrix, fmt.Errorf("unknown metric %+v", metr) //
	case "counter":
		if val, ok := memorial.Countmetr[metr.ID]; ok {
			out := int64(val)
			metrix.Delta = &out
			break
		}
		return metrix, fmt.Errorf("unknown metric %+v", metr)
	default:
		return metrix, fmt.Errorf("wrong type %s", metr.MType)
	}
	return metrix, nil
}

// --- from []Metrics to memory Storage
func (memorial *MemoryStorageStruct) PutAllMetrics(ctx context.Context, metras *[]Metrics) error {
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()

	for _, metr := range *metras {
		switch metr.MType {
		case "gauge":
			memorial.Gaugemetr[metr.ID] = gauge(*metr.Value)
		case "counter":
			if _, ok := memorial.Countmetr[metr.ID]; ok {
				memorial.Countmetr[metr.ID] += counter(*metr.Delta)
				continue
			}
			memorial.Countmetr[metr.ID] = counter(*metr.Delta)
		default:
			log.Printf("wrong metric type %s\n", metr.MType)
		}
	}
	return nil
}

// ----- from Memory Storage to []Metrics
func (memorial *MemoryStorageStruct) GetAllMetrics(ctx context.Context) (*[]Metrics, error) {

	memorial.Mutter.RLock()
	defer memorial.Mutter.RUnlock()

	metras := []Metrics{}

	for nam, val := range memorial.Countmetr {
		out := int64(val)
		metr := Metrics{ID: nam, MType: "counter", Delta: &out}
		metras = append(metras, metr)
	}
	for nam, val := range memorial.Gaugemetr {
		out := float64(val)
		metr := Metrics{ID: nam, MType: "gauge", Value: &out}
		metras = append(metras, metr)
	}
	return &metras, nil
}

// -------------------------------  FILERs ------------------------------------------
type MStorJSON struct {
	Gaugemetr map[string]models.Gauge
	Countmetr map[string]models.Counter
}

func (memorial *MemoryStorageStruct) UnmarshalMS(data []byte) error {
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
func (memorial *MemoryStorageStruct) MarshalMS() ([]byte, error) {
	buf := new(bytes.Buffer)
	memorial.Mutter.RLock()
	err := json.NewEncoder(buf).Encode(MStorJSON{
		Gaugemetr: memorial.Gaugemetr,
		Countmetr: memorial.Countmetr,
	})
	memorial.Mutter.RUnlock()
	return append(buf.Bytes(), '\n'), err
}

func (memorial *MemoryStorageStruct) LoadMS(fnam string) error {
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
func (memorial *MemoryStorageStruct) SaveMS(fnam string) error {
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

func (memorial *MemoryStorageStruct) Saver(fnam string, storeInterval int) error {
	for {
		time.Sleep(time.Duration(storeInterval) * time.Second)
		err := memorial.SaveMS(fnam)
		if err != nil {
			return fmt.Errorf("save err %v", err)
		}
	}
}
func (memorial *MemoryStorageStruct) Ping(ctx context.Context) error {
	return fmt.Errorf(" Skotobaza closed")
}
