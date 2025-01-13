package middles

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

type responseData struct {
	status int
	size   int
}
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func WithLogging(origFunc func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	loggedFunc := func(w http.ResponseWriter, r *http.Request) {
		//	start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		origFunc(&lw, r) // обслуживание оригинального запроса

		//	duration := time.Since(start)
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Println("cannot initialize zap")
		}
		defer logger.Sync()
		sugar = *logger.Sugar()
		// sugar.Infoln(
		// 	"uri", r.RequestURI,
		// 	"method", r.Method,
		// 	"status", responseData.status, // получаем перехваченный код статуса ответа
		// 	"duration", duration,
		// 	"size", responseData.size, // получаем перехваченный размер ответа
		// )
	}
	return loggedFunc
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

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
			req = newReq
		}

		next.ServeHTTP(rwr, req)
	})
}

/*
curl localhost:8080/update/ -H "Content-Type":"application/json" -d "{\"type\":\"gauge\",\"id\":\"nam\",\"value\":77}"
*/
