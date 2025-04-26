package handlera

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func (suite *TstHandlers) Test_02SimplePut() {
	// gauge
	val := "12345.678"
	urlVarsPut := map[string]string{"metricType": "gauge", "metricName": "AllocG", "metricValue": val}
	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w := httptest.NewRecorder()
	PutMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(val, bb)
	// counter
	delt := "54321"
	urlVarsPut = map[string]string{"metricType": "counter", "metricName": "AllocC", "metricValue": delt}
	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	bb = w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(delt, bb)

	// wrong metric type
	urlBadType := map[string]string{"metricType": "gauge1", "metricName": "Alloc", "metricValue": val}
	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlBadType)
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

}

func (suite *TstHandlers) Test_03SimpleGet() {
	// gauge
	val := "12345.678"
	urlVarsPut := map[string]string{"metricType": "gauge", "metricName": "AllocG"}
	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w := httptest.NewRecorder()
	GetMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(val, bb)
	// counter
	delt := "54321"
	urlVarsPut = map[string]string{"metricType": "counter", "metricName": "AllocC"}
	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w = httptest.NewRecorder()
	GetMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	bb = w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(delt, bb)
}

// Test_SimpleCounterTwice - проверка того, что тип counter суммирует значения
func (suite *TstHandlers) Test_04SimpleCounterTwice() {

	urlVarsPut := map[string]string{"metricType": "counter", "metricName": "cunt", "metricValue": "44"}

	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut) // 44
	w := httptest.NewRecorder()
	PutMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal("44", bb)

	urlVarsPut = map[string]string{"metricType": "counter", "metricName": "cunt", "metricValue": "46"}
	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut) // 46
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	bb = w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal("90", bb) // 44 + 46

}

func (suite *TstHandlers) Test_05badPost() {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		urla string
		want want
	}{
		{
			name: "Right case",
			urla: "/",
			want: want{
				code:        http.StatusNotFound,
				response:    `{"status":"StatusNotFound"}`,
				contentType: "text/html",
			},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			request := httptest.NewRequest(http.MethodPost, tt.urla, nil)
			w := httptest.NewRecorder()
			BadPost(w, request)
			res := w.Result()
			defer res.Body.Close()

			suite.Assert().Equal(tt.want.code, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)

			suite.Require().NoError(err)
			suite.Assert().JSONEq(tt.want.response, string(resBody))
			suite.Assert().Equal(tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func (suite *TstHandlers) Test_06GetAllMetricsHandler() {
	request := httptest.NewRequest(http.MethodGet, "/wtf", nil)
	w := httptest.NewRecorder()
	GetAllMetricsHandler(w, request)
	res := w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

	val := "12345"
	//delt := 54321
	urlVarsPut := map[string]string{"metricType": "counter", "metricName": "Alloca", "metricValue": val}

	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w = httptest.NewRecorder()
	PutMetric(w, request)

	// so add gauge metric
	//suite.Test_SimplePutGet()

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	GetAllMetricsHandler(w, request)
	res = w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
}
func (suite *TstHandlers) Test_07dbPing() {
	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	w := httptest.NewRecorder()
	DBPinger(w, request)
	res := w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
}
