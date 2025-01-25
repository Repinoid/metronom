package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/models"
	"gorono/internal/privacy"
	"io"
	"log"
	"net/http"
	"strings"
)

// /value/ handler
func GetJSONMetric(rwr http.ResponseWriter, req *http.Request) {
	rwr.Header().Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		sugar.Debugf("io.ReadAll %+v\n", err)
		return
	}
	defer req.Body.Close()

	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		sugar.Debugf("json.Unmarshal %+v err %+v\n", metr, err)
		return
	}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metr)
	if err == nil { // if ништяк
		rwr.WriteHeader(http.StatusOK)
		json.NewEncoder(rwr).Encode(metr)
		return
	}

	//sugar.Debugf("after inter.GetMetric %+v err %+v\n", metr, err)

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
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным типом метрики или значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	defer req.Body.Close()

	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}

	if !models.IsMetricsOK(metr) {
		rwr.WriteHeader(http.StatusBadRequest)
		sugar.Debugf("bad Metric %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.PutMetricWrapper(inter.PutMetric)(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		sugar.Debugf("PutMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	metrix := Metrics{ID: metr.ID, MType: metr.MType}
	metr, err = basis.GetMetricWrapper(inter.GetMetric)(ctx, &metrix) //inter.GetMetric(ctx, &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		sugar.Debugf("GetMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(metr)

	if storeInterval == 0 {
		_ = inter.SaveMS(fileStorePath)
	}
}

func Buncheras(rwr http.ResponseWriter, req *http.Request) {
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! io.ReadAll(req.Body) err %+v\n", err)
		return
	}
	defer req.Body.Close()

	buf := bytes.NewBuffer(telo)
	metras := []models.Metrics{}
	err = json.NewDecoder(buf).Decode(&metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! bunch decode  err %+v\n", err)
		return
	}
	err = basis.PutAllMetricsWrapper(inter.PutAllMetrics)(ctx, &metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! Put   err %+v\n", err)
		return
	}

	if key != "" {
		keyB := md5.Sum([]byte(key)) //[]byte(key)
		toencrypt, _ := json.Marshal(&metras)

		coded, err := privacy.EncryptB2B(toencrypt, keyB[:])
		if err != nil {
			sugar.Debugf("encrypt   err %+v\n", err)
			return
		}
		ha := privacy.MakeHash(nil, coded, keyB[:])
		haHex := hex.EncodeToString(ha)
		rwr.Header().Add("HashSHA256", haHex)
	}

	rwr.WriteHeader(http.StatusOK)
	json.NewEncoder(rwr).Encode(&metras)
}

func CryptoHandleDecoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {

		if haInHeader := req.Header.Get("HashSHA256"); haInHeader != "" {
			telo, err := io.ReadAll(req.Body)
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			defer req.Body.Close()

			keyB := md5.Sum([]byte(key)) //[]byte(key)
			ha := privacy.MakeHash(nil, telo, keyB[:])
			haHex := hex.EncodeToString(ha)

			log.Printf("%s from KEY %s\n%s from Header\n", haHex, key, haInHeader)

			if haHex != haInHeader {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"wrong hash":"%s"}`, haInHeader)
				return
			}
			telo, err = privacy.DecryptB2B(telo, keyB[:])
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewBuffer(telo))
			if err != nil {
				io.WriteString(rwr, err.Error())
				return
			}
			for name := range req.Header {
				hea := req.Header.Get(name)
				newReq.Header.Add(name, hea)
			}
			req = newReq
		}
		next.ServeHTTP(rwr, req)
	})
}
