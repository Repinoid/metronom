package handlera

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
	"gorono/internal/privacy"
	"io"
	"log"
	"net/http"
	"strings"
)

type Metrics = memos.Metrics

// /value/ handler
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

	metr := Metrics{}
	err = json.Unmarshal([]byte(telo), &metr)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest) // с некорректным  значением возвращать http.StatusBadRequest.
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		models.Sugar.Debugf("json.Unmarshal %+v err %+v\n", metr, err)
		return
	}
	err = basis.CommonMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
	if err == nil { // if ништяк
		rwr.WriteHeader(http.StatusOK)
		json.NewEncoder(rwr).Encode(metr) // return marshalled metric
		return
	}

	//models.Sugar.Debugf("after models.Inter.GetMetric %+v err %+v\n", metr, err)

	if strings.Contains(err.Error(), "unknown metric") {
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
		models.Sugar.Debugf("bad Metric %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.CommonMetricWrapper(models.Inter.PutMetric)(req.Context(), &metr, nil)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		models.Sugar.Debugf("PutMetricWrapper %+v\n", metr)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	err = basis.CommonMetricWrapper(models.Inter.GetMetric)(req.Context(), &metr, nil)
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

func Buncheras(rwr http.ResponseWriter, req *http.Request) {
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! io.ReadAll(req.Body) err %+v\n", err)
		return
	}
	defer req.Body.Close()

	buf := bytes.NewBuffer(telo)
	metras := []models.Metrics{}
	err = json.NewDecoder(buf).Decode(&metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! bunch decode  err %+v\n", err)
		return
	}

	err = basis.CommonMetricWrapper(models.Inter.PutAllMetrics)(req.Context(), nil, &metras)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		models.Sugar.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! Put   err %+v\n", err)
		return
	}

	if models.Key != "" {
		keyB := md5.Sum([]byte(models.Key)) //[]byte(key)
		toencrypt, _ := json.Marshal(&metras)

		coded, err := privacy.EncryptB2B(toencrypt, keyB[:])
		if err != nil {
			models.Sugar.Debugf("encrypt   err %+v\n", err)
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

		if haInHeader := req.Header.Get("HashSHA256"); haInHeader != "" { // если есть ключ переопределить req
			telo, err := io.ReadAll(req.Body)
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			defer req.Body.Close()

			keyB := md5.Sum([]byte(models.Key)) //[]byte(key)
			ha := privacy.MakeHash(nil, telo, keyB[:])
			haHex := hex.EncodeToString(ha)

			log.Printf("%s from KEY %s\n%s from Header\n", haHex, models.Key, haInHeader)

			if haHex != haInHeader { // несовпадение хешей вычисленного по ключу и переданного в header
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
			for name := range req.Header { // cкопировать поля header
				hea := req.Header.Get(name)
				newReq.Header.Add(name, hea)
			}
			req = newReq
		}
		next.ServeHTTP(rwr, req)
	})
}
