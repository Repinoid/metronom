package handlera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

// test GzipHandleEncoder middleware on different metrics functions
func (suite *TstHandlers) Test_gzipPutGet() {
	// //initForTests()
	// InitServer()
	type want struct {
		code     int
		response string
		//		err      error
	}
	controlMetric := models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr[float64](78)}
	cmMarshalled, _ := json.Marshal(controlMetric)
	controlMetric1 := models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr[float64](77)}
	cmMarshalled1, _ := json.Marshal(controlMetric1)

	bunch := []models.Metrics{controlMetric, controlMetric1}
	bunchOnMarsh, _ := json.Marshal(bunch)

	tests := []struct {
		name            string
		AcceptEncoding  string
		ContentEncoding string
		ContentType     string
		want            want
		metr            models.Metrics
		metras          []models.Metrics
		function        func(http.ResponseWriter, *http.Request) //func4test
	}{
		// {
		// 	name:            "GET unknown",
		// 	AcceptEncoding:  "gzip",
		// 	ContentEncoding: "",
		// 	ContentType:     "application/json",
		// 	function:        GetJSONMetric,
		// 	metr:            controlMetric,
		// 	want: want{
		// 		code:     http.StatusOK,
		// 		response: `{"status":"StatusNotFound"}`, // Metric not exist yet
		// 	},
		// },
		{
			name:            "BUNCH",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        Buncheras,
			metras:          bunch,
			want: want{
				code:     http.StatusOK,
				response: string(bunchOnMarsh),
			},
		},
		{
			name:            "PutJSONMetric AcceptEncoding",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        PutJSONMetric,
			metr:            controlMetric,
			want: want{
				code:     http.StatusOK,
				response: string(cmMarshalled),
			},
		},
		{
			name:            "GET After PUT",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        GetJSONMetric,
			metr:            controlMetric,
			want: want{
				code:     http.StatusOK,
				response: string(cmMarshalled),
			},
		},

		{
			name:            "NO ENCODINg",
			AcceptEncoding:  "",
			ContentEncoding: "gzip",
			//ContentType:     "application/json",
			function: thecap,
			metr:     controlMetric1,
			want: want{
				code:     http.StatusOK,
				response: string(cmMarshalled1),
			},
		},
		{
			name:            "THECAP AcceptEncoding:  \"gzip\"",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        PutJSONMetric,
			metr: models.Metrics{
				MType: "gaug",
				ID:    "Alloc",
				Value: middlas.Ptr[float64](77),
			},
			want: want{
				code:     http.StatusOK,
				response: `{"status":"StatusBadRequest"}`,
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var march []byte
			if tt.name == "BUNCH" {
				march, _ = json.Marshal(tt.metras) // []Metrics
			} else {
				march, _ = json.Marshal(tt.metr)
			}

			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(march))
			w := httptest.NewRecorder()

			request.Header.Set("Accept-Encoding", tt.AcceptEncoding)
			request.Header.Set("Content-Encoding", tt.ContentEncoding)
			request.Header.Set("Content-Type", tt.ContentType)

			fu := tt.function
			hfunc := http.HandlerFunc(fu)          // make handler from function
			hh := middlas.GzipHandleEncoder(hfunc) // оборачиваем в мидлварь который зипует
			hh.ServeHTTP(w, request)               // запускаем handler

			res := w.Body // ответ
			var telo []byte

			if tt.AcceptEncoding == "gzip" {
				//		if tt.ContentEncoding == "gzip" {
				unpak, err := middlas.UnpackFromGzip(res) // распаковка
				if err != nil {
					log.Printf("UnpackFromGzip %+v\n", err)
				}
				telo, err = io.ReadAll(unpak)
				if err != nil {
					log.Printf("AcceptEncoding == \"gzip\" io.ReadAll %+v\n", err)
				}
			}
			if tt.ContentEncoding == "gzip" {
				var err error
				telo, err = io.ReadAll(res)
				if err != nil {
					log.Printf("ContentEncoding == \"gzip\" io.ReadAll %+v\n", err)
				}
			}
			suite.Assert().JSONEq(tt.want.response, string(telo))

		})
	}
}

// хандлер для теста - что пришло, то и ушло
func thecap(rwr http.ResponseWriter, req *http.Request) { 
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	defer req.Body.Close()
	rwr.Write(telo)
}

// go test ./... -v -coverpkg=./...
