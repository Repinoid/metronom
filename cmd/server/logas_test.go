package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func Test_gzipHandlePLUG(t *testing.T) {
	type want struct {
		code     int
		response string
		//		err      error
	}
	tests := []struct {
		name            string
		AcceptEncoding  string
		ContentEncoding string
		ContentType     string
		want            want
		metr            Metrics
	}{
		{
			name:            "AcceptEncoding:  \"gzip\"",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: Pfloat64(77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"status":"StatusBadRequest"}`,
			},
		},
		{
			AcceptEncoding:  "",
			ContentEncoding: "gzip",
			//ContentType:     "application/json",
			name: "No encoding",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: Pfloat64(77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"status":"StatusBadRequest"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			march, _ := json.Marshal(tt.metr)

			// if tt.ContentEncoding == "gzip" {
			// 	march, _ = pack2gzip(march)
			// }
			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(march))
			w := httptest.NewRecorder()

			request.Header.Set("Accept-Encoding", tt.AcceptEncoding)
			request.Header.Set("Content-Encoding", tt.ContentEncoding)
			request.Header.Set("Content-Type", tt.ContentType)

			hfunc := http.HandlerFunc(thecap)
			hh := gzipHandleEncoder(hfunc)
			hh.ServeHTTP(w, request)

			res := w.Body

			if tt.AcceptEncoding == "gzip" {
				//		if tt.ContentEncoding == "gzip" {
				u, err := unpackFromGzip(res)
				if err != nil {
					fmt.Println(err)
				}
				telo, _ := io.ReadAll(u)
				ma, _ := json.Marshal(tt.metr)
				assert.JSONEq(t, string(ma), string(telo))
			}
			if tt.ContentEncoding == "gzip" {
				telo, _ := io.ReadAll(res)
				ma, _ := json.Marshal(tt.metr)
				assert.JSONEq(t, string(ma), string(telo))
				//	assert.Equal(t, tt.want.response, res.)
			}

		})
	}
}

func thecap(rwr http.ResponseWriter, req *http.Request) {
	//	rwr.Header().Set("Content-Type", "application/json")
	//	req.Header.Set("Content-Type", "application/json")

	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	rwr.Write(telo)

}

/*
func Test_getjsonvalue(t *testing.T) {
	type want struct {
		code     int
		response string
//		err      error
	}
	tests := []struct {
		name            string
		AcceptEncoding  string
		ContentEncoding string
		ContentType     string
		want            want
		metr            Metrics
	}{
		{
			name:            "Get RIGHT",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: Pfloat64(77.77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"value":77.77,"type":"gauge","id":"Alloc"}`,
			},
		},
		{
			name:            "Wrong value",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc2",
				Value: Pfloat64(77),
			},
			want: want{
				code:     http.StatusNotFound,
				response: `{"status":"StatusNotFound"}`,
			},
		},
		{
			name:            "Get counter value",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "counter",
				ID:    "PollCounter",
				Delta: Pint64(77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"delta":77,"type":"counter","id":"PollCounter"}`,
			},
		},
	}
	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}
	memStor.count["PollCounter"] = 77
	memStor.gau["Alloc"] = 77.77

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			march, _ := json.Marshal(tt.metr)

			// if tt.ContentEncoding == "gzip" {
			// 	march, _ = pack2gzip(march)
			// }
			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(march))
			w := httptest.NewRecorder()

			// request.Header.Set("Accept-Encoding", tt.AcceptEncoding)
			// request.Header.Set("Content-Encoding", tt.ContentEncoding)
			// request.Header.Set("Content-Type", tt.ContentType)

			hfunc := http.HandlerFunc(getJSONMetric)
			hh := gzipHandleEncoder(hfunc)
			hh.ServeHTTP(w, request)

			res := w.Body
			telo, _ := io.ReadAll(res)
			//ma, _ := json.Marshal(tt.metr)
			assert.JSONEq(t, tt.want.response, string(telo))
			assert.Equal(t, tt.want.code, w.Result().StatusCode)
			//assert.Equal(t, tt.want.code, w.Code)

		})
	}
}
func Test_updater(t *testing.T) {
	type want struct {
		code     int
		response string
//		err      error
	}
	tests := []struct {
		name            string
		AcceptEncoding  string
		ContentEncoding string
		ContentType     string
		want            want
		metr            Metrics
	}{
		{
			name:            "put_RIGHT",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: Pfloat64(77.77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"value":77.77,"type":"gauge","id":"Alloc"}`,
			},
		},
		{
			name:            "Wrong_type",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "gaug",
				ID:    "Alloc2",
				Value: Pfloat64(77),
			},
			want: want{
				code:     http.StatusBadRequest,
				response: `{"status":"StatusBadRequest"}`,
			},
		},
		{
			name:            "SET_counter_value",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			metr: Metrics{
				MType: "counter",
				ID:    "PollCounter",
				Delta: Pint64(77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"delta":77,"type":"counter","id":"PollCounter"}`,
			},
		},
	}
	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}
	//	memStor.count["PollCounter"] = 77
	//	memStor.gau["Alloc"] = 77.77

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			march, _ := json.Marshal(tt.metr)

			if tt.ContentEncoding == "gzip" {
				march, _ = pack2gzip(march)
			}
			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(march))
			w := httptest.NewRecorder()

			request.Header.Set("Accept-Encoding", tt.AcceptEncoding)
			request.Header.Set("Content-Encoding", tt.ContentEncoding)
			request.Header.Set("Content-Type", tt.ContentType)

			hfunc := http.HandlerFunc(treatJSONMetric)
			hh := gzipHandleEncoder(hfunc)
			hh.ServeHTTP(w, request)

			res := w.Body

			if tt.AcceptEncoding == "gzip" {
				u, err := unpackFromGzip(res)
				if err != nil {
					fmt.Println(err)
					//	assert.Equal(t, tt.want.response, u)
				}
				telo, _ := io.ReadAll(u)
				ma, _ := json.Marshal(tt.metr)
				assert.JSONEq(t, string(ma), string(telo))
			}
			if tt.ContentEncoding == "gzip" {
				telo, _ := io.ReadAll(res)
				ma, _ := json.Marshal(tt.metr)
				assert.JSONEq(t, string(ma), string(telo))
				//	assert.Equal(t, tt.want.response, res.)
			}

			//ma, _ := json.Marshal(tt.metr)
			//			assert.JSONEq(t, tt.want.response, string(telo))
			//			assert.Equal(t, tt.want.code, w.Result().StatusCode)
			//assert.Equal(t, tt.want.code, w.Code)

		})
	}
}
*/
