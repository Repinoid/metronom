package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostMetric(t *testing.T) {
	type want struct {
		ret error
	}
	tests := []struct {
		name                                string
		metricType, metricName, metricValue string
		want                                want // пока не надо
	}{
		{
			name:       "err1",
			metricType: "gaug", metricName: "Alloc", metricValue: "77.77",
			want: want{
				ret: nil,
			},
		},
		{
			name:       "err2",
			metricType: "gauge", metricName: "Allo", metricValue: "77.77",
			want: want{
				ret: nil,
			},
		},
		{
			name:       "err3",
			metricType: "gauge", metricName: "Alloc", metricValue: "77g.77",
			want: want{
				ret: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := postMetric(tt.metricType, tt.metricName, tt.metricValue)
			assert.Error(t, res)
		})
	}
}
