package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/models"
	"io"
	"net/http"
)

func getJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	var inta int64
	var flo float64
	metr := Metrics{Value: &flo, Delta: &inta}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metrix := Metrics{ID: metr.ID, MType: metr.MType}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metrix)
	if err != nil {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metr)
}

func treatJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}

	if !models.IsMetricsOK(metr) {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, "bad metric %+v\n", metr)
		return
	}
	err = basis.PutMetricWrapper(inter.PutMetric)(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		return
	}
	metrix := Metrics{ID: metr.ID, MType: metr.MType}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metrix) //inter.GetMetric(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metr)

	if storeInterval == 0 {
		_ = inter.SaveMS(fileStorePath)
	}
}

func buncheras(rwr http.ResponseWriter, req *http.Request) {
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
