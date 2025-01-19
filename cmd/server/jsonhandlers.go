package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/models"
	"io"
	"log"
	"net/http"
	"strings"
)

func GetJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metr)
	if err == nil { // if ништяк
		rwr.WriteHeader(http.StatusOK)
		json.NewEncoder(rwr).Encode(metr)
		return
	}
	if strings.Contains(err.Error(), "unknown metric") {
		//rwr.WriteHeader(444) // неизвестной метрики сервер должен возвращать http.StatusNotFound.
		rwr.WriteHeader(http.StatusNotFound) // неизвестной метрики сервер должен возвращать http.StatusNotFound.
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики http.StatusBadRequest.
	fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
}

func PutJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	if !models.IsMetricsOK(metr) {
		rwr.WriteHeader(http.StatusBadRequest)
		log.Printf("bad Metric %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.PutMetricWrapper(inter.PutMetric)(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		log.Printf("PutMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metrix := Metrics{ID: metr.ID, MType: metr.MType}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metrix) //inter.GetMetric(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		log.Printf("GetMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metr)

	if storeInterval == 0 {
		_ = memStor.SaveMS(fileStorePath)
	}
}

func Buncheras(rwr http.ResponseWriter, req *http.Request) {
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	buf := bytes.NewBuffer(telo)
	metras := []models.Metrics{}
	err = json.NewDecoder(buf).Decode(&metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	err = basis.PutAllMetricsWrapper(inter.PutAllMetrics)(ctx, &metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(&metras)
}
