package middlas

import (
	"bytes"
	"fmt"
	"gorono/internal/models"
	"gorono/internal/privacy"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

// хандлер для теста - что пришло, то и ушло
func thecap(rwr http.ResponseWriter, req *http.Request) {
	telo, err := io.ReadAll(req.Body)
	if err != nil {
		rwr.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rwr, `{"status":"StatusBadRequest"}`)
		return
	}
	defer req.Body.Close()
	rwr.Write(telo)
}

func (suite *TstMid) Test_zips() {

	msg := []byte("progon")

	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(msg))
	w := httptest.NewRecorder()

	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("Content-Encoding", "")
	request.Header.Set("Content-Type", "application/json")

	hfunc := http.HandlerFunc(thecap) // make handler from function
	hh := GzipHandleEncoder(hfunc)    // оборачиваем в мидлварь который зипует
	hh.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()

	telo := w.Body

	request = httptest.NewRequest(http.MethodPost, "/value/", telo)
	w = httptest.NewRecorder()

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	hfunc = http.HandlerFunc(thecap) // make handler from function
	hh = GzipHandleDecoder(hfunc)    // оборачиваем в мидлварь который unзипует
	hh.ServeHTTP(w, request)

	res = w.Result()
	defer res.Body.Close()

	teloOrig := w.Body

	suite.Assert().True(bytes.Equal(msg, teloOrig.Bytes()))

}

func (suite *TstMid) Test_cryptaDecoder() {

	// any file for testing bytes
	acc, err := os.ReadFile("../../cmd/agent/agent.exe")
	suite.Assert().NoError(err)

	// Public Key File
	pkb, err := os.ReadFile("../../cmd/agent/cert.pem")
	suite.Assert().NoError(err)

	// Private key file
	privK, err := os.ReadFile("../../cmd/server/privateKey.pem")
	suite.Assert().NoError(err)
	models.PrivateKey = string(privK)

	cipherByte, err := privacy.Encrypt(acc, pkb)
	suite.Assert().NoError(err)

	request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(cipherByte))
	w := httptest.NewRecorder()

	hfunc := http.HandlerFunc(thecap) // make handler from function
	hh := CryptoHandleDecoder(hfunc)  // оборачиваем в мидлварь
	hh.ServeHTTP(w, request)

	res := w.Result()
	defer res.Body.Close()

	telo := w.Body

	suite.Assert().True(bytes.Equal(acc, telo.Bytes()))

	a := Ptr(4.5)
	suite.Assert().Equal(*a, 4.5)
}
