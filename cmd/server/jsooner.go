package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func pack2gzip(data2pack []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.ModTime = time.Now()
	_, err := zw.Write(data2pack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Write %w ", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("gzip.NewWriter.Close %w ", err)
	}
	return buf.Bytes(), nil
}
func unpackFromGzip(data2unpack io.Reader) (io.Reader, error) {
	gzipReader, err := gzip.NewReader(data2unpack)
	if err != nil {
		return nil, fmt.Errorf("gzip.NewReader %w ", err)
	}
	if err := gzipReader.Close(); err != nil {
		return nil, fmt.Errorf("zr.Close %w ", err)
	}
	return gzipReader, nil
}

func getJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")
	//req.Header.Set("Content-Type", "application/json")
	//rwr.WriteHeader(http.StatusBadRequest)
	//return
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	//	inta := int64(0)
	//	floata := float64(0)
	metra := Metrics{} //Delta: &inta, Value: &floata}
	err = json.Unmarshal([]byte(telo), &metra)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		//	fmt.Fprintf(rwr, "%v", telo)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	switch metra.MType {
	case "counter":
		var cunt counter
		if memStor.getCounterValue(metra.ID, &cunt) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, `{"status":"StatusNotFound"}`)
			//			json.NewEncoder(rwr).Encode(metra)
			return
		}
		delt := int64(cunt)
		metra.Delta = &delt

	//	json.NewEncoder(rwr).Encode(metra)
	//	return

	// resp, err := json.Marshal(metra)
	// if err != nil {
	// 	http.Error(rwr, err.Error(), http.StatusInternalServerError)
	// 	fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
	// 	return
	// }
	// rwr.Write(resp)
	case "gauge":
		var gaaga gauge
		if memStor.getGaugeValue(metra.ID, &gaaga) != nil {
			rwr.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rwr, `{"status":"StatusNotFound"}`)
			//			break
			return
		}
		flo := float64(gaaga)
		metra.Value = &flo

	//	json.NewEncoder(rwr).Encode(metra)

	// resp, err := json.Marshal(metra)
	// if err != nil {
	// 	http.Error(rwr, err.Error(), http.StatusInternalServerError)
	// 	fmt.Fprintf(rwr, `{"status":"StatusInternalServerError"}`)
	// 	return
	// }
	// rwr.Write(resp)
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	json.NewEncoder(rwr).Encode(metra) // common
}

func treatJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")
	req.Header.Set("Content-Type", "application/json")

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
		json.NewEncoder(rwr).Encode(metra)

		// resp, err := json.Marshal(metra)
		// if err != nil {
		// 	http.Error(rwr, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// rwr.Write(resp)
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
		json.NewEncoder(rwr).Encode(metra)

		// resp, err := json.Marshal(metra)
		// if err != nil {
		// 	http.Error(rwr, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// rwr.Write(resp)
	default:
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
}
