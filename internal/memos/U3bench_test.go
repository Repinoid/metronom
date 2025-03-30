package memos

import (
	"encoding/json"
	"gorono/internal/models"
	"testing"
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

// func BenchmarkNewDecoder(b *testing.B) {
// 	b.StopTimer()
// 	metricsBunch := []models.Metrics{m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2, m1, m2}
// 	metricsOut := []models.Metrics{}
// 	ma, _ := json.Marshal(metricsBunch)
// 	buf := bytes.NewBuffer(ma)
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		//json.Unmarshal(ma, &metricsOut)
// 		json.NewDecoder(buf).Decode(&metricsOut)
// 	}
// 	// fmt.Println(metricsOut[0], len(metricsOut))
// }

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

// go tool pprof -http=":9090" cpuo
// go test -bench .  -cpuprofile=cpuo
