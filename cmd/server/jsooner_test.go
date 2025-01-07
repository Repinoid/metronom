package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Pointer[T any](v T) *T {
	return &v
}
func Pint64(arg int64) *int64 {
	return &arg
}
func Pfloat64(arg float64) *float64 {
	return &arg
}
func Test_treatJSONMetric(t *testing.T) {

	//v77 := 77.77
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
		metr Metrics
	}{
		{
			name: "Wrong type",
			metr: Metrics{
				MType: "gaug",
				ID:    "Alloc",
				Value: Pfloat64(77),
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    `{"status":"StatusBadRequest"}`,
				contentType: "application/json",
			},
		},
		{
			name: "Right gauge case",
			metr: Metrics{
				MType: "gauge",
				ID:    "Alloc",
				Value: Pfloat64(77),
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"id":"Alloc", "type":"gauge", "value":77}`,
				contentType: "application/json",
			},
		},
		{
			name: "Right counter case",
			metr: Metrics{
				MType: "counter",
				ID:    "Alloca",
				Delta: Pint64(77),
			},
			want: want{
				code:        http.StatusOK,
				response:    `{"id":"Alloca", "type":"counter", "delta":77}`,
				contentType: "application/json",
			},
		},
	}
	memStor = MemStorage{
		Gaugemetr:   make(map[string]gauge),
		Countmetr: make(map[string]counter),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urla := "/update/"
			march, _ := json.Marshal(tt.metr)
			request := httptest.NewRequest(http.MethodPost, urla, bytes.NewBuffer(march))

			w := httptest.NewRecorder()
			treatJSONMetric(w, request)
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
