package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
)

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
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		origFunc(&lw, r) // обслуживание оригинального запроса

		duration := time.Since(start)

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", responseData.size, // получаем перехваченный размер ответа
		)
	}
	return loggedFunc
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(respon http.ResponseWriter, claim *http.Request) {
		rwr := respon
		req := claim
		if strings.Contains(claim.Header.Get("Accept-Encoding"), "gzip") &&
			(strings.Contains(claim.Header.Get("Content-Type"), "application/json") ||
				strings.Contains(claim.Header.Get("Content-Type"), "text/html")) {
			respon.Header().Set("Content-Encoding", "gzip")        //
			req.Header.Set("Content-Encoding", "gzip")             //
			gz, err := gzip.NewWriterLevel(respon, gzip.BestSpeed) // compressing
			if err != nil {
				io.WriteString(respon, err.Error())
				return
			}
			defer gz.Close()
			rwr = gzipWriter{ResponseWriter: respon, Writer: gz}
		}
		if strings.Contains(claim.Header.Get("Content-Encoding"), "gzip") {
			rwr.Header().Set("Content-Encoding", "")
			req.Header.Set("Content-Encoding", "")
			gzipReader, err := gzip.NewReader(claim.Body) // decompressing
			if err != nil {
				io.WriteString(respon, err.Error())
				return
			}
			newReq, err := http.NewRequest(claim.Method, claim.URL.String(), gzipReader)
			if err != nil {
				io.WriteString(respon, err.Error())
				return
			}
			req = newReq
		}
		next.ServeHTTP(rwr, req)
	})
}
