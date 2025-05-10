// пакет работы с Memory Storage
package memos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"gorono/internal/models"
)

// структура для хранения метрик - "база данных в памяти"
type MemoryStorageStruct struct {
	Gaugemetr map[string]models.Gauge
	Countmetr map[string]models.Counter
	Mutter    *sync.RWMutex
}

// создание MemoryStorage
func InitMemoryStorage() *MemoryStorageStruct {
	var mtx sync.RWMutex
	memStor := MemoryStorageStruct{
		Gaugemetr: make(map[string]models.Gauge),
		Countmetr: make(map[string]models.Counter),
		Mutter:    &mtx,
	}
	return &memStor
}

// записать метрику в базу в памяти
func (memorial *MemoryStorageStruct) PutMetric(ctx context.Context, metr *models.Metrics, gag *[]models.Metrics) error {
	if !IsMetricOK(*metr) {
		return fmt.Errorf("bad metric %+v", metr)
	}
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()
	switch metr.MType {
	case "gauge":
		memorial.Gaugemetr[metr.ID] = models.Gauge(*metr.Value)
	case "counter":
		memorial.Countmetr[metr.ID] += models.Counter(*metr.Delta)
	}
	return nil
}

// получить метрику из MemStorage
func (memorial *MemoryStorageStruct) GetMetric(ctx context.Context, metr *models.Metrics, gag *[]models.Metrics) error {
	memorial.Mutter.RLock() // <---- MUTEX
	defer memorial.Mutter.RUnlock()
	switch metr.MType {
	case "gauge":
		if val, ok := memorial.Gaugemetr[metr.ID]; ok {
			out := float64(val)
			metr.Value = &out
			break
		}
		return fmt.Errorf("unknown metric %+v", metr) //
	case "counter":
		if val, ok := memorial.Countmetr[metr.ID]; ok {
			out := int64(val)
			metr.Delta = &out
			break
		}
	default:
		return fmt.Errorf("wrong type %s", metr.MType)
	}
	return nil
}

// from []models.Metrics to memory Storage
func (memorial *MemoryStorageStruct) PutAllMetrics(ctx context.Context, gag *models.Metrics, metras *[]models.Metrics) error {
	memorial.Mutter.Lock()
	defer memorial.Mutter.Unlock()

	for _, metr := range *metras {
		switch metr.MType {
		case "gauge":
			memorial.Gaugemetr[metr.ID] = models.Gauge(*metr.Value)
		case "counter":
			if _, ok := memorial.Countmetr[metr.ID]; ok {
				memorial.Countmetr[metr.ID] += models.Counter(*metr.Delta)
				continue
			}
			memorial.Countmetr[metr.ID] = models.Counter(*metr.Delta)
		default:
			return fmt.Errorf("wrong type %s", metr.MType)
		}
	}
	return nil
}

// from Memory Storage to []models.Metrics
func (memorial *MemoryStorageStruct) GetAllMetrics(ctx context.Context, gag *models.Metrics, meS *[]models.Metrics) error {

	memorial.Mutter.RLock()
	defer memorial.Mutter.RUnlock()

	var metras []models.Metrics

	for nam, val := range memorial.Countmetr {
		out := int64(val)
		metr := models.Metrics{ID: nam, MType: "counter", Delta: &out}
		metras = append(metras, metr)
	}
	for nam, val := range memorial.Gaugemetr {
		out := float64(val)
		metr := models.Metrics{ID: nam, MType: "gauge", Value: &out}
		metras = append(metras, metr)
	}
	*meS = metras
	return nil
}

// проверка базы данных, когда активен Memory Storage. Открывает базу, пингует и закрывает
func (memorial *MemoryStorageStruct) Ping(ctx context.Context, dbEndPoint string) error {
	db, err := pgx.Connect(ctx, dbEndPoint)
	if err != nil {
		log.Println(" Skotobaza closed !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return err
	}
	defer db.Close(ctx) // при пинге из активного memory storage соединяемся с базой с нуля, пингуем и закрываем
	err = db.Ping(ctx)
	if err != nil {
		log.Println(" Skotobaza closed !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		return err
	}
	return nil
}

// получить текущее имя интерфейса Inter
func (memorial *MemoryStorageStruct) GetName() string {
	return "Memorial"
}

// структура для анмаршаллинга -  MemoryStorageStruct без мьютекса
type MStorJSON struct {
	Gaugemetr map[string]models.Gauge
	Countmetr map[string]models.Counter
}

// decode Memstorage
func (memorial *MemoryStorageStruct) UnmarshalMS(data []byte) error {
	memor := MStorJSON{
		Gaugemetr: make(map[string]models.Gauge),
		Countmetr: make(map[string]models.Counter),
	}
	buf := bytes.NewBuffer(data)
	memorial.Mutter.Lock()
	err := json.NewDecoder(buf).Decode(&memor)
	memorial.Gaugemetr = memor.Gaugemetr
	memorial.Countmetr = memor.Countmetr
	memorial.Mutter.Unlock()
	return err
}

// Encode Memstorage
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

// загрузить метрики в Memstorage из файла
func (memorial *MemoryStorageStruct) LoadMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	defer phil.Close()

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

// сохранить метрики в Memstorage в файл
func (memorial *MemoryStorageStruct) SaveMS(fnam string) error {
	phil, err := os.OpenFile(fnam, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("file %s Open error %v", fnam, err)
	}
	defer phil.Close()

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

// для горутины - сохранение метрик через storeInterval секунд
func (memorial *MemoryStorageStruct) Saver(ctx context.Context, fnam string, storeInterval int, wg *sync.WaitGroup) error {
	defer wg.Done()
	ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)

	for {
		select {

		case <-ctx.Done():
			log.Println("Запись метрик в файл остановлена")
			return errors.New("Saver остановлен по сигналу")

		case <-ticker.C:
			err := memorial.SaveMS(fnam)
			if err != nil {
				return fmt.Errorf("save err %v", err)
			}
		}
	}
}

func (memorial *MemoryStorageStruct) Close() {
	log.Println("MS Closed")
}

// check if Metric has correct fields
func IsMetricOK(metr models.Metrics) bool {
	if (metr.MType != "gauge" && metr.MType != "counter") ||
		(metr.MType == "counter" && metr.Delta == nil) ||
		(metr.MType == "gauge" && metr.Value == nil) ||
		(metr.Delta != nil && metr.Value != nil) {
		return false
	}
	return true
}
