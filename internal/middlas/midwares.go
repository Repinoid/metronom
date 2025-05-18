// пакет мидлварей
package middlas

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
func IpcidrCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rwr http.ResponseWriter, req *http.Request) {

		var ipnet *net.IPNet
		// если есть СИДР - проверяем вхождение в подсеть переданного агентом хеадера X-Real-IP
		if models.Cidr != "" {
			// третий параметр - ошибка, проверена при инициализации сервера
			_, ipnet, _ = net.ParseCIDR(models.Cidr)

			// get X-Real-IP from agent request header
			agentIP, err := ipFromHeader(req)
			if err != nil {
				rwr.WriteHeader(http.StatusForbidden)
				// зaпись ошибки в формате JSON
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				models.Sugar.Debugf("нет хеадера X-Real-IP %+v\n", err)
				io.WriteString(rwr, err.Error())
				return
			}
			aIP := net.ParseIP(agentIP.String())
			// если aIP (который X-Real-IP от агента) НЕ входит в сабнет CIDR (ipnet)
			if !ipnet.Contains(aIP) {
				rwr.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(rwr, `{"%s NOT in CIDR ":"%s"}`, aIP, ipnet)
				models.Sugar.Debugf("%s NOT in CIDR (ipnet) %s\n", aIP, ipnet)
				return
			}
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
			teloDecr, err := privacy.Decrypt(telo, []byte(models.PrivateKey))
			if err != nil {
				rwr.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rwr, `{"Error":"%v"}`, err)
				return
			}
			newReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewBuffer(teloDecr))
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

	})
}

// ipFromHeader получает заголовок X-Real-IP, в котором должен содержаться IP-адрес хоста агента.
func ipFromHeader(r *http.Request) (net.IP, error) {
	// смотрим заголовок запроса X-Real-IP
	ipStr := r.Header.Get("X-Real-IP")
	// парсим ip
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// если заголовок X-Real-IP пуст, пробуем X-Forwarded-For
		// этот заголовок содержит адреса отправителя и промежуточных прокси
		// в виде 203.0.113.195, 70.41.3.18, 150.172.238.178
		ips := r.Header.Get("X-Forwarded-For")
		// разделяем цепочку адресов
		ipStrs := strings.Split(ips, ",")
		// интересует только первый
		ipStr = ipStrs[0]
		// парсим
		ip = net.ParseIP(ipStr)
	}
	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}
	return ip, nil
}

func OurInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	var ipnet *net.IPNet
	// если есть СИДР - проверяем вхождение в подсеть переданного агентом хеадера X-Real-IP
	if models.Cidr != "" {
		// третий параметр - ошибка, проверена при инициализации сервера
		_, ipnet, _ = net.ParseCIDR(models.Cidr)

		agentIP := ""
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("X-Real-IP")
			if len(values) == 0 {
				models.Sugar.Debugf("нет хеадера X-Real-IP\n")
				return nil, status.Error(codes.NotFound, "нет X-Real-IP")
			}
			agentIP = values[0]
		}

		aIP := net.ParseIP(agentIP)
		// если aIP (который X-Real-IP от агента) НЕ входит в сабнет CIDR (ipnet)
		if !ipnet.Contains(aIP) {
			models.Sugar.Debugf("%s НЕ входит в сабнет %s", agentIP, ipnet)
			return nil, status.Errorf(codes.PermissionDenied, "%s НЕ входит в сабнет %s", agentIP, ipnet)
		}
	}

	return handler(ctx, req)
}
