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
			name:       "err1",
			metricType: "hague", metricName: "Alloc", metricValue: "77.77",
			want: want{
				ret: "wrong metric type",
			},
		},
		{
			name:       "err2",
			metricType: "counter", metricName: "Allo", metricValue: "77.77",
			want: want{
				ret: "ParseInt",
			},
		},
		{
			name:       "err3",
			metricType: "gauge", metricName: "Alloc", metricValue: "77g.77",
			want: want{
				ret: "ParseFloat",
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
