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
	mutter    *sync.RWMutex
}
type Metrics = models.Metrics

func (memorial MemoryStorageStruct) PutMetric(ctx context.Context, metr *Metrics) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()
	switch metr.MType {
	case "gauge":
		memorial.Gaugemetr[metr.ID] = models.Gauge(*metr.Value)
	case "counter":
		memorial.Countmetr[metr.ID] += models.Counter(*metr.Delta)
	default:
		return fmt.Errorf("wrong metric %+v", metr)
	}
	return nil
}

func (memorial MemoryStorageStruct) GetMetric(ctx context.Context, metr *Metrics) (Metrics, error) {
	memorial.mutter.RLock() // <---- MUTEX
	defer memorial.mutter.RUnlock()
	switch metr.MType {
	case "gauge":
		if val, ok := memorial.Gaugemetr[metr.ID]; ok {
			out := float64(val)
			metr.Value = &out
			break
		}
		return *metr, fmt.Errorf("no metric %+v", metr)
	case "counter":
		if val, ok := memorial.Countmetr[metr.ID]; ok {
			out := int64(val)
			metr.Delta = &out
			break
		}
		return *metr, fmt.Errorf("no metric %+v", metr)
	default:
		return *metr, fmt.Errorf("wrong metric %+v", metr)
	}
	return *metr, nil
}

// --- from []Metrics to memory Storage
func (memorial MemoryStorageStruct) PutAllMetrics(ctx context.Context, metras *[]Metrics) error {
	memorial.mutter.Lock()
	defer memorial.mutter.Unlock()

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
func (memorial MemoryStorageStruct) GetAllMetrics(ctx context.Context) (*[]Metrics, error) {

	memorial.mutter.RLock()
	defer memorial.mutter.RUnlock()

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

func UnmarshalMS(memorial *MemoryStorageStruct, data []byte) error {
	memor := MStorJSON{
		Gaugemetr: make(map[string]gauge),
		Countmetr: make(map[string]counter),
	}
	buf := bytes.NewBuffer(data)
	memorial.mutter.Lock()
	err := json.NewDecoder(buf).Decode(&memor)
	memorial.Gaugemetr = memor.Gaugemetr
	memorial.Countmetr = memor.Countmetr
	memorial.mutter.Unlock()
	return err
}
func MarshalMS(memorial *MemoryStorageStruct) ([]byte, error) {
	buf := new(bytes.Buffer)
	memorial.mutter.RLock()
	err := json.NewEncoder(buf).Encode(MStorJSON{
		Gaugemetr: memorial.Gaugemetr,
		Countmetr: memorial.Countmetr,
	})
	memorial.mutter.RUnlock()
	return append(buf.Bytes(), '\n'), err
}

func (memorial MemoryStorageStruct) LoadMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	reader := bufio.NewReader(phil)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("file %s Read error %v", fnam, err)
	}
	err = UnmarshalMS(&memorial, data)
	if err != nil {
		return fmt.Errorf(" Memstorage UnMarshal error %v", err)
	}
	return nil
}
func (memorial MemoryStorageStruct) SaveMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	march, err := MarshalMS(&memorial)
	if err != nil {
		return fmt.Errorf(" Memstorage Marshal error %v", err)
	}
	_, err = phil.Write(march)
	if err != nil {
		return fmt.Errorf("file %s Write error %v", fnam, err)
	}
	return nil
}

func (memorial MemoryStorageStruct) Saver(fnam string, storeInterval int) error {
	for {
		time.Sleep(time.Duration(storeInterval) * time.Second)
		err := memorial.SaveMS(fnam)
		if err != nil {
			return fmt.Errorf("save err %v", err)
		}
	}
}
func (memorial MemoryStorageStruct) Ping(ctx context.Context) error {
	return fmt.Errorf(" Skotobaza closed")
}
