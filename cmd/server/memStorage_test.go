package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_getCounterValue(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name  string
		args  args
		noErr bool
	}{
		{
			name: "Right case",
			args: args{
				name:  "metricName",
				value: "105",
			},
			noErr: true,
		},
		{
			name: "noname",
			args: args{
				name:  "noname",
				value: "105",
			},
			noErr: false,
		},
		{
			name: "wrongvalue",
			args: args{
				name:  "metricName",
				value: "78",
			},
			noErr: false,
		},
	}
	memStor = MemStorage{
		gau:   make(map[string]gauge),
		count: make(map[string]counter),
	}
	memStor.addCounter("metricName", counter(50))
	memStor.addCounter("metricName", counter(55)) // 105 = 50+55
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var str string
			err := memStor.getCounterValue(tt.args.name, &str)
			erra := err == nil && tt.args.value == str // corrent both name & value
			assert.Equal(t, erra, tt.noErr)
		})
	}
}
