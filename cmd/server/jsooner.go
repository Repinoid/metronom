package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func getJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metra := Metrics{}
	err = json.Unmarshal([]byte(telo), &metra)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	switch metra.MType {
	case "counter":
		//err = memStor.getCounterValue(metricName, &val)
		var cunt counter
		if memStor.getCounterValue(metra.ID, &cunt) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		delt := int64(cunt)
		metra.Delta = &delt
		resp, err := json.Marshal(metra)
		if err != nil {
			http.Error(rwr, err.Error(), http.StatusInternalServerError)
			return
		}
		rwr.Write(resp)
	case "gauge":
		//		err = memStor.getGaugeValue(metricName, &val)
		var gaaga gauge
		if memStor.getGaugeValue(metra.ID, &gaaga) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		flo := float64(gaaga)
		metra.Value = &flo
		resp, err := json.Marshal(metra)
		if err != nil {
			http.Error(rwr, err.Error(), http.StatusInternalServerError)
			return
		}
		rwr.Write(resp)
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(rwr, nil)
		return
	}
}

func treatJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metra := Metrics{}
	json.Unmarshal([]byte(telo), &metra)
	metricType := metra.MType
	metricName := metra.ID
	metricValue := metra.Value
	metricDelta := metra.Delta

	//	fmt.Fprint(rwr, string(telo), metra)
	if metricValue == nil && metricDelta == nil {
		rwr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(rwr, `{"status":"StatusNotFound"}`)
		return
	}
	switch metricType {
	case "counter":
		if metricDelta == nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addCounter(metricName, counter(*metricDelta))
		// get new value from memstorage
		var cunt counter
		if memStor.getCounterValue(metra.ID, &cunt) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		*metra.Delta = int64(cunt)

		resp, err := json.Marshal(metra)
		if err != nil {
			http.Error(rwr, err.Error(), http.StatusInternalServerError)
			return
		}
		rwr.Write(resp)
	case "gauge":
		if metricValue == nil {
			rwr.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
			return
		}
		memStor.addGauge(metricName, gauge(*metricValue))
		// get new value from memstorage
		var gaaga gauge
		if memStor.getGaugeValue(metra.ID, &gaaga) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, nil)
			return
		}
		*metra.Value = float64(gaaga)

		resp, err := json.Marshal(metra)
		if err != nil {
			http.Error(rwr, err.Error(), http.StatusInternalServerError)
			return
		}
		rwr.Write(resp)
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
}
