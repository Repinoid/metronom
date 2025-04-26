// пакет хендлеров - обработчиков запросов на сервер
package handlera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
)

// GetJSONMetric - возвращает метрику по запросу методом POST.
// router.HandleFunc("/value/", handlera.GetJSONMetric).Methods("POST")
func GetJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debugf("io.ReadAll %+v\n", err)
		return
	}
	defer req.Body.Close()

	metr := models.Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debugf("json.Unmarshal %+v err %+v\n", metr, err)
		return
	}
	err = basis.RetryMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
	if err == nil { // if ништяк
		rwr.WriteHeader(http.StatusOK)
		json.NewEncoder(rwr).Encode(metr) // return marshalled metric
		return
	}

	if strings.Contains(err.Error(), "unknown metric") {
		rwr.WriteHeader(http.StatusNotFound) // неизвестной метрики сервер должен возвращать http.StatusNotFound.
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики http.StatusBadRequest.
	fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
}

// PutJSONMetric - размещает метрику по запросу методом POST
// router.HandleFunc("/update/", handlera.PutJSONMetric).Methods("POST")
func PutJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	defer req.Body.Close()

	metr := models.Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	if !memos.IsMetricOK(metr) {
		rwr.WriteHeader(http.StatusBadRequest)
		models.Sugar.Debugf("bad Metric %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.RetryMetricWrapper(models.Inter.PutMetric)(req.Context(), &metr, nil)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		models.Sugar.Debugf("PutMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.RetryMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		models.Sugar.Debugf("GetMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metr) // return marshalled metric

	if models.StoreInterval == 0 {
		_ = models.Inter.SaveMS(models.FileStorePath)
	}
}

// Buncheras - размещает слайс метрик. В т.ч. от Агента.
// router.HandleFunc("/updates/", handlera.Buncheras).Methods("POST")
func Buncheras(rwr http.ResponseWriter, req *http.Request) {
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf("!!!!! io.ReadAll(req.Body) err %+v\n", err)
		return
	}
	defer req.Body.Close()

	buf := bytes.NewBuffer(telo)
	metras := []models.Metrics{}
	err = json.NewDecoder(buf).Decode(&metras)

	//metras, err := memos.MetrixUnMarhal(telo) // own json decoder

	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf("bunch decode  err %+v\n", err)
		return
	}

	err = basis.RetryMetricWrapper(models.Inter.PutAllMetrics)(req.Context(), nil, &metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf(" Put   err %+v\n", err)
		return
	}

	// models.Key задаётся переменной откружения или аргументом -k
	// if models.Key != "" {
	// 	//хеш-функция — MD5 из пакета https://pkg.go.dev/crypto/md5. Bозвращает хеш длиной 16 байт.
	// 	keyB16 := md5.Sum([]byte(models.Key))

	// 	keyB := keyB16[:]                            // keyB16 [16]byte => keyB []byte
	// 	coded, err := privacy.EncryptB2B(telo, keyB) // кодируем telo => coded
	// 	if err != nil {
	// 		models.Sugar.Debugf("encrypt   err %+v\n", err)
	// 		return
	// 	}
	// 	ha := privacy.MakeHash(nil, coded, keyB)
	// 	haHex := hex.EncodeToString(ha)       // EncodeToString returns the hexadecimal encoding of src.
	// 	rwr.Header().Add("HashSHA256", haHex) // добавляем в заголовок хэш
	// }

	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metras) // return marshalled metrics slice
}
