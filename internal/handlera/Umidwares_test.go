package handlera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"net/http"
	"net/http/httptest"

	"gorono/internal/middlas"
	"gorono/internal/models"
)

// test GzipHandleEncoder middleware on different metrics functions
func (suite *TstHandlers) Test_01gzipPutGet() {
	type want struct {
		code     int
		response string
		//		err      error
	}
	badMetric := models.Metrics{MType: "gaug", ID: "Alloc", Value: middlas.Ptr[float64](78)}
	badcmMarshalled, _ := json.Marshal(badMetric)
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
			name:            "BAD PutJSONMetric AcceptEncoding",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        PutJSONMetric,
			metr:            badMetric,
			want: want{
				code:     http.StatusBadRequest,
				response: string(badcmMarshalled),
			},
		},
		{
			name:            "BAD GetJSONMetric AcceptEncoding",
			AcceptEncoding:  "gzip",
			ContentEncoding: "",
			ContentType:     "application/json",
			function:        GetJSONMetric,
			metr:            badMetric,
			want: want{
				code:     http.StatusBadRequest,
				response: string(badcmMarshalled),
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
			function:        thecap,
			metr:            controlMetric1,
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
				code:     http.StatusBadRequest,
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

			res := w.Result()

			if tt.AcceptEncoding == "gzip" {
				//		if tt.ContentEncoding == "gzip" {
				unpak, err := middlas.UnpackFromGzip(w.Body) // распаковка
				if err != nil {
					log.Printf("UnpackFromGzip %+v\n", err)
				}
				_, err = io.ReadAll(unpak)
				if err != nil {
					log.Printf("AcceptEncoding == \"gzip\" io.ReadAll %+v\n", err)
				}
			}
			if tt.ContentEncoding == "gzip" {
				var err error
				_, err = io.ReadAll(w.Body)
				if err != nil {
					log.Printf("ContentEncoding == \"gzip\" io.ReadAll %+v\n", err)
				}
			}
			//			_= telo
			suite.Assert().Equal(tt.want.code, w.Result().StatusCode)
			
			res.Body.Close()
			//suite.Assert().JSONEq(tt.want.response, string(telo))

		})
	}
}

func (suite *TstHandlers) Test_01BadBunch() {
	request := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer([]byte("hzwhat")))
	w := httptest.NewRecorder()
	Buncheras(w, request)
	res := w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	request = httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer([]byte("hzwhat")))
	w = httptest.NewRecorder()
	GetJSONMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	// controlMetric := models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr[float64](78)}
	// //cmMarshalled, _ := json.Marshal(controlMetric)
	// controlMetric1 := models.Metrics{MType: "countera", ID: "Alloc", Value: middlas.Ptr[float64](77)}
	// //cmMarshalled1, _ := json.Marshal(controlMetric1)

	// bunch := []models.Metrics{controlMetric, controlMetric1}
	// bunchOnMarsh, _ := json.Marshal(bunch)

	// request = httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(bunchOnMarsh))
	// w = httptest.NewRecorder()
	// Buncheras(w, request)
	// res = w.Result()
	// defer res.Body.Close()
	// suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

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
