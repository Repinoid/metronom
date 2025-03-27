package middlas

import (
	"bytes"
	"fmt"
	"gorono/internal/models"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func (suite *TstMid) Test01MiddlasSugared() {

	tstBuf := []byte("qwerty")
	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(tstBuf))
	w := httptest.NewRecorder()

	fu := thecap
	hfunc := http.HandlerFunc(fu) // make handler from function

	hh := WithLogging(hfunc)
	hh.ServeHTTP(w, request)
	res := w.Body
	telo, err := io.ReadAll(res)
	suite.Assert().NoError(err)
	suite.Assert().Equal(tstBuf, telo)

}
func (suite *TstMid) Test01MiddlasNoSugar() {

	tstBuf := []byte("qwerty")
	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(tstBuf))
	w := httptest.NewRecorder()

	fu := thecap
	hfunc := http.HandlerFunc(fu) // make handler from function

	models.Logger, _ = zap.NewDevelopment()
	hh := NoSugarLogging(hfunc)
	hh.ServeHTTP(w, request)
	res := w.Body
	telo, err := io.ReadAll(res)
	suite.Assert().NoError(err)
	suite.Assert().Equal(tstBuf, telo)

}

func BenchmarkSugared(b *testing.B) {
	b.StopTimer()

	tstBuf := []byte("qwerty")
	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(tstBuf))
	w := httptest.NewRecorder()

	fu := thecap
	hfunc := http.HandlerFunc(fu) // make handler from function
	hh := WithLogging(hfunc)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hh.ServeHTTP(w, request)
	}
}
func BenchmarkNoSugar(b *testing.B) {
	b.StopTimer()
	models.Logger, _ = zap.NewDevelopment()
	tstBuf := []byte("qwerty")
	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(tstBuf))
	w := httptest.NewRecorder()
	fu := thecap
	hfunc := http.HandlerFunc(fu) // make handler from function
	hh := NoSugarLogging(hfunc)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		hh.ServeHTTP(w, request)
	}
}

func thecap(rwr http.ResponseWriter, req *http.Request) { // хандлер для теста - что пришло, то и ушло
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	defer req.Body.Close()
	rwr.Write(telo)
}
