// пакет мидлварей
package middlas

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"gorono/internal/models"
	"gorono/internal/privacy"
)

// responseData структура для ZAP logger.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter структура для ZAP logger
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

// метод middleware ZAP logger, захватываем размер
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// метод middleware ZAP logger, захватываем statusCode в заголовке
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging ZAP log with SUGAR
func WithLogging(next http.Handler) http.Handler {
	loggedFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)
		models.Sugar.Infoln(
			"uri", r.URL.Path, // какой именно эндпоинт был дернут
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}

	return http.HandlerFunc(loggedFunc)
}

// NoSugarLogging ZAP log no sugar mode
func NoSugarLogging(next http.Handler) http.Handler {
	loggedFunc := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)
		models.Logger.Info("NoSug ",
			zap.String("uri", r.URL.Path), // какой именно эндпоинт был дернут
			zap.String("method", r.Method),
			zap.Int("status", responseData.status), // получаем перехваченный код статуса ответа
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size), // получаем перехваченный размер ответа
		)
	}
	return http.HandlerFunc(loggedFunc)
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GzipHandleEncoder middleware упаковки
func GzipHandleEncoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {
		isTypeOK := strings.Contains(req.Header.Get("Content-Type"), "application/json") ||
			strings.Contains(req.Header.Get("Content-Type"), "text/html") ||
			strings.Contains(req.Header.Get("Accept"), "application/json") ||
			strings.Contains(req.Header.Get("Accept"), "text/html")

		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && isTypeOK {
			rwr.Header().Set("Content-Encoding", "gzip") //
			gz := gzip.NewWriter(rwr)                    // compressing
			defer gz.Close()
			rwr = gzipWriter{ResponseWriter: rwr, Writer: gz}
		}
		next.ServeHTTP(rwr, req)
	})
}

// GzipHandleDecoder middleware распаковки
func GzipHandleDecoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {

		if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(req.Body) // decompressing
			if err != nil {
				io.WriteString(rwr, err.Error())
				return
			}
			newReq, err := http.NewRequest(req.Method, req.URL.String(), gzipReader)
			if err != nil {
				io.WriteString(rwr, err.Error())
				return
			}
			for name := range req.Header {
				hea := req.Header.Get(name)
				newReq.Header.Add(name, hea)
			}
			req = newReq
		}

		next.ServeHTTP(rwr, req)
	})
}

// CryptoHandleDecoder middleware раскодировки криптографии
func CryptoHandleDecoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {
		// если указан  models.Key файл с private key
		if models.Key != "" {
			telo, err := io.ReadAll(req.Body)
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			defer req.Body.Close()
			// models.PrivateKey - содержимое файла в models.Key
			telo, err = privacy.Decrypt(telo, []byte(models.PrivateKey))
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewBuffer(telo))
			if err != nil {
				io.WriteString(rwr, err.Error())
				return
			}
			// cкопировать поля header
			for name := range req.Header {
				hea := req.Header.Get(name)
				newReq.Header.Add(name, hea)
			}
			// переопределяем request
			req = newReq
		}
		next.ServeHTTP(rwr, req)

		// if haInHeader := req.Header.Get("HashSHA256"); haInHeader != "" { // если есть ключ переопределить req
		// 	telo, err := io.ReadAll(req.Body)
		// 	if err != nil {
		// 		rwr.WriteHeader(http.StatusBadRequest)
		// 		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		// 		return
		// 	}
		// 	defer req.Body.Close()

		// 	keyB := md5.Sum([]byte(models.Key)) //[]byte(key)
		// 	ha := privacy.MakeHash(nil, telo, keyB[:])
		// 	haHex := hex.EncodeToString(ha)

		// 	//			log.Printf("%s from KEY %s\n%s from Header\n", haHex, models.Key, haInHeader)

		// 	if haHex != haInHeader { // несовпадение хешей вычисленного по ключу и переданного в header
		// 		rwr.WriteHeader(http.StatusBadRequest)
		// 		fmt.Fprintf(rwr, `{"wrong hash":"%s"}`, haInHeader)
		// 		return
		// 	}
		// 	telo, err = privacy.DecryptB2B(telo, keyB[:])
		// 	if err != nil {
		// 		rwr.WriteHeader(http.StatusBadRequest)
		// 		fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
		// 		return
		// 	}
		// 	newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewBuffer(telo))
		// 	if err != nil {
		// 		io.WriteString(rwr, err.Error())
		// 		return
		// 	}
		// 	for name := range req.Header { // cкопировать поля header
		// 		hea := req.Header.Get(name)
		// 		newReq.Header.Add(name, hea)
		// 	}
		// 	req = newReq
		// }
		// next.ServeHTTP(rwr, req)
	})
}
