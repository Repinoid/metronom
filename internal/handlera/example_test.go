package handlera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func ExamplePutMetric() {

	memStor := memos.InitMemoryStorage()
	models.Inter = memStor

	val := "2026.0308"
	urlVarsPut := map[string]string{"metricType": "gauge", "metricName": "gaaga", "metricValue": val}

	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut) // в горилле только так можно передать переменные URL
	w := httptest.NewRecorder()
	PutMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()
	fmt.Println(bb)

	// Output:
	// 2026.0308

}
func ExamplePutMetric_badInt() {

	memStor := memos.InitMemoryStorage()
	models.Inter = memStor

	val := "2026.0308"
	urlVarsPut := map[string]string{"metricType": "counter", "metricName": "cunt", "metricValue": val}

	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w := httptest.NewRecorder()
	PutMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()

	fmt.Println(bb)

	// Output:
	// {"status":"StatusBadRequest"}

}

func ExampleGetMetric() {

	ExamplePutMetric()

	urlVarsPut := map[string]string{"metricType": "gauge", "metricName": "gaaga"}
	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w := httptest.NewRecorder()
	GetMetric(w, request)
	res := w.Result()
	res.Body.Close()

	bb := w.Body.String()
	fmt.Println(bb)
	// вывод - 2 раза по val := "2026.0308" из ExamplePutMetric. Первый output - результат от PutMetric

	// Output:
	// 2026.0308
	// 2026.0308

}

func ExamplePutJSONMetric() {

	memStor := memos.InitMemoryStorage()
	models.Inter = memStor

	controlMetric := models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr(78.87)}
	march, _ := json.Marshal(controlMetric)
	request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(march))
	w := httptest.NewRecorder()

	PutJSONMetric(w, request)

	res := w.Body
	telo, err := io.ReadAll(res)
	if err != nil {
		return
	}
	metr := models.Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)

	fmt.Println(metr.MType, metr.ID, *metr.Value)

	// Output:
	// gauge Alloc 78.87

}

func ExampleGetJSONMetric() {

	memStor := memos.InitMemoryStorage()
	models.Inter = memStor

	controlMetric := models.Metrics{MType: "gauge", ID: "Alloc1", Value: middlas.Ptr(78.77)}
	march, _ := json.Marshal(controlMetric)
	request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(march))
	w := httptest.NewRecorder()
	PutJSONMetric(w, request)

	controlMetric = models.Metrics{MType: "gauge", ID: "Alloc1"}
	march, _ = json.Marshal(controlMetric)
	request = httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(march))
	w = httptest.NewRecorder()
	GetJSONMetric(w, request)

	res := w.Body
	telo, err := io.ReadAll(res)
	if err != nil {
		return
	}
	metr := models.Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)

	fmt.Println(metr.MType, metr.ID, *metr.Value)

	// Output:
	// gauge Alloc1 78.77

}

func ExampleBuncheras() {

	memStor := memos.InitMemoryStorage()
	models.Inter = memStor

	controlMetric := models.Metrics{MType: "gauge", ID: "Alloc2", Value: middlas.Ptr(78.76)}
	mm := []models.Metrics{controlMetric}
	march, _ := json.Marshal(mm)
	request := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(march))
	w := httptest.NewRecorder()

	Buncheras(w, request)

	res := w.Body
	telo, err := io.ReadAll(res)
	if err != nil {
		return
	}
	metr := []models.Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)

	fmt.Println(metr[0].MType, metr[0].ID, *metr[0].Value)

	// Output:
	// gauge Alloc2 78.76

}
