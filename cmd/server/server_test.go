package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getMetric(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		want want
		name string
		metr map[string]string
	}{
		{
			name: "Right case",
			metr: map[string]string{
				"metricType":  "gauge",
				"metricName":  "Alloc",
				"metricValue": "77.77",
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"status":"StatusOK"}`,
				contentType: "text/html",
			},
		},
		{
			name: "wrong type",
			metr: map[string]string{
				"metricType":  "gaug",
				"metricName":  "Alloc",
				"metricValue": "77.77",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `{"status":"StatusBadRequest"}`,
				contentType: "text/html",
			},
		},
		{
			name: "Wrong value",
			metr: map[string]string{
				"metricType":  "gauge",
				"metricName":  "Alloc",
				"metricValue": "77.a77",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `{"status":"StatusBadRequest"}`,
				contentType: "text/html",
			},
		},
		{
			name: "no value",
			metr: map[string]string{
				"metricType": "gauge",
				"metricName": "Alloc",
				//	"metricValue": "77.a77",
			},
			want: want{
				code:        http.StatusNotFound,
				response:    `{"status":"StatusNotFound"}`,
				contentType: "text/html",
			},
		},
	}
	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urla := "/update/" + tt.metr["metricType"] + "/" + tt.metr["metricName"] + "/" + tt.metr["metricValue"]
			request := httptest.NewRequest(http.MethodPost, urla, nil)
			request = mux.SetURLVars(request, tt.metr)

			w := httptest.NewRecorder()
			treatMetric(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.JSONEq(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
