package middlas

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.uber.org/zap"

	"gorono/internal/models"
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

func (suite *TstMid) Test07zips() {

	fromFile, err := os.ReadFile("../../cmd/server/server.exe")
	suite.Assert().NoError(err)

	bts, err := Pack2gzip(fromFile)
	suite.Assert().NoError(err)

	unp, err := UnpackFromGzip(bytes.NewReader(bts))
	suite.Assert().NoError(err)

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(unp)
	suite.Assert().NoError(err)

	suite.Assert().True(bytes.Equal(fromFile, buf.Bytes()))

}
