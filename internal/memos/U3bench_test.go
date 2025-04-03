package memos

import (
	"encoding/json"
	"testing"

	"gorono/internal/models"
)

func BenchmarkOwnUnMarsh(b *testing.B) {
	b.StopTimer()
	metricsBunch := []models.Metrics{m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2}
	//metricsBunch := []models.Metrics{m1, m2}
	ma, _ := json.Marshal(metricsBunch)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		MetrixUnMarhal(ma)
	}
}


func BenchmarkJSONUnMarshal(b *testing.B) {
	b.StopTimer()
	metricsBunch := []models.Metrics{m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2}
	metricsOut := []models.Metrics{}
	ma, _ := json.Marshal(metricsBunch)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(ma, &metricsOut)
	}
}

