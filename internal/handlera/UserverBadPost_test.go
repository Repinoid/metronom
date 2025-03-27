package handlera

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

func (suite *TstHandlers) Test_badPost() {
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

func (suite *TstHandlers) Test_BadGetAllMetricsHandler() {
	request := httptest.NewRequest(http.MethodGet, "/wtf", nil)
	w := httptest.NewRecorder()
	GetAllMetricsHandler(w, request)
	res := w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (suite *TstHandlers) Test_SimplePutGet() {
	val := "12345.678"
	//delt := 54321
	urlVarsPut := map[string]string{"metricType": "gauge", "metricName": "Alloc", "metricValue": val}

	request := httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut)
	w := httptest.NewRecorder()
	PutMetric(w, request)
	res := w.Result()
	defer res.Body.Close()
	bb := w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(val, bb)

	urlVarsGet := map[string]string{"metricType": "gauge", "metricName": "Alloc"}

	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsGet)
	w = httptest.NewRecorder()
	GetMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	bb = w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal(val, bb)

	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsGet)
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusNotFound, res.StatusCode)

	urlBadType := map[string]string{"metricType": "gauge1", "metricName": "Alloc", "metricValue": val}
	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlBadType)
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	suite.Assert().Equal(http.StatusBadRequest, res.StatusCode)

}

func (suite *TstHandlers) Test_SimpleCounterTwice() {

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

	request = httptest.NewRequest(http.MethodGet, "/pofiguchto", nil)
	request = mux.SetURLVars(request, urlVarsPut) // 44
	w = httptest.NewRecorder()
	PutMetric(w, request)
	res = w.Result()
	defer res.Body.Close()
	bb = w.Body.String()
	suite.Assert().Equal(http.StatusOK, res.StatusCode)
	suite.Assert().Equal("88", bb) // 44 + 44

}

// func (suite *TstHandlers) Test_SimplePutGetGauge() {

// 	metr := models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr(777.77)}
// 	urla := fmt.Sprintf("/update/%s/%s/%g", metr.MType, metr.ID, *metr.Value)
// 	request := httptest.NewRequest(http.MethodGet, urla, nil)
// 	w := httptest.NewRecorder()
// 	PutMetric(w, request)
// 	res := w.Body
// 	telo, err := io.ReadAll(res)
// 	suite.Require().Equal(err != nil, false)
// 	suite.Require().Equal("777.77", string(telo))

// }
