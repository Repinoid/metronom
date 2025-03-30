package memos

import (
	"encoding/json"
	"gorono/internal/middlas"
	"gorono/internal/models"
)

var m1 = models.Metrics{MType: "gauge", ID: "gaager", Value: middlas.Ptr(444.44)}
var m2 = models.Metrics{MType: "counter", ID: "cunter", Delta: middlas.Ptr[int64](444)}

func (suite *TstMemo) Test01MetrixUnMarshal() {
	metricsBunch := []models.Metrics{m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2}

	ma, _ := json.Marshal(metricsBunch)

	tests := []struct {
		name       string
		metroB     []models.Metrics
		marchalled []byte
		wantErr    bool
	}{
		{
			name:       "Bad metr, byte suffix added",
			marchalled: append(ma, byte(66)),
			wantErr:    true,
		},
		{
			name:       "Nice UnMarch",
			marchalled: ma,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_, err := MetrixUnMarhal(tt.marchalled)
			suite.Require().Equal(tt.wantErr, err != nil)
		})
	}
}
