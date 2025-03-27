package handlera

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"
	"io"
	"net/http"
	"net/http/httptest"
)

func (suite *TstHandlers) Test_cryptas() {

	tests := []struct {
		name        string
		key         string
		inputString []byte
		metr        models.Metrics
		function    func(http.ResponseWriter, *http.Request) //func4test
	}{
		{
			name:        "crypto Right",
			key:         "keykey",
			inputString: []byte("whtatToSend"),
			//			function:    handlera.PutMetric,
			function: thecap,
		},
		{
			name:        "crypto Right2",
			key:         "key\"key\"dfgdfgdfg___6567567#$%$#",
			inputString: []byte("whtatToSenddfgdfgdfg#$%#$%#$%dfgdfgdfgdfg\"dfgdfgdfgdfg"),
			function:    thecap,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {

			keyB := md5.Sum([]byte(tt.key))

			coded, err := privacy.EncryptB2B([]byte(tt.inputString), keyB[:])
			suite.Assert().NoError(err)
			ha := privacy.MakeHash(nil, coded, keyB[:])
			haHex := hex.EncodeToString(ha)

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(coded)) // post crypted tt.inputString
			request.Header.Add("HashSHA256", haHex)

			w := httptest.NewRecorder()

			models.Key = tt.key // for CryptoHandleDecoder
			fu := tt.function
			hfunc := http.HandlerFunc(fu)            // make handler from function
			hh := middlas.CryptoHandleDecoder(hfunc) // оборачиваем в мидлварь который расшифрует
			hh.ServeHTTP(w, request)

			res := w.Body
			telo, err := io.ReadAll(res)
			suite.Assert().NoError(err)
			suite.Assert().Equal(telo, tt.inputString)

		})
	}
}
