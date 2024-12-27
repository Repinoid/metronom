package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostMetric(t *testing.T) {
	type want struct {
		ret string
	}
	tests := []struct {
		name                                string
		metricType, metricName, metricValue string
		want                                want // пока не надо
	}{
		{
			name:       "Ok Gauge",
			metricType: "gauge", metricName: "Alloc", metricValue: "77.77",
			want: want{
				ret: "", //nil,
			},
		},
		{
			name:       "wrong metric type",
			metricType: "hague", metricName: "Alloc", metricValue: "77.77",
			want: want{
				ret: "wrong metric type", //fmt.Errorf("wrong metric type"),
			},
		},
		{
			name:       "wrong counter value",
			metricType: "counter", metricName: "Allo", metricValue: "77.77",
			want: want{
				ret: "wrong counter value", // fmt.Errorf("wrong counter value"),
			},
		},
		{
			name:       "wrong gauge value",
			metricType: "gauge", metricName: "Alloc", metricValue: "77g.77",
			want: want{
				ret: "wrong gauge value", //fmt.Errorf("wrong gauge value"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := postMetric(tt.metricType, tt.metricName, tt.metricValue)
			assert.ErrorContains(t, res, tt.want.ret)
		})
	}
}
