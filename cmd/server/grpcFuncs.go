package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	_ "net/http/pprof"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "gorono/cmd/proto"

	"gorono/internal/basis"
	"gorono/internal/models"
)

// MetricServer поддерживает все необходимые методы сервера.
type MetricServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName> для совместимости с будущими версиями
	pb.UnimplementedMetricServer
}

func (s *MetricServer) AddBunch(ctx context.Context, in *pb.Bunch) (*pb.BunchResponse, error) {
	var response pb.BunchResponse

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

	// from []*pb.GMetr to []models.Metrics
	metras := make([]models.Metrics, len(in.Meters))
	for i, v := range in.Meters {
		metras[i].MType = v.GetMType()
		metras[i].ID = v.GetID()
		metras[i].Delta = &v.Delta
		metras[i].Value = &v.Value
	}

	err := basis.RetryMetricWrapper(models.Inter.PutAllMetrics)(ctx, nil, &metras)
	if err != nil {
		models.Sugar.Debugf("grpc PutAllMetrics   err %+v\n", err)
		return nil, status.Errorf(codes.Unimplemented, "grpc PutAllMetrics err %+v code %v\n", err, status.Code(err))
	}
	response.OutData = fmt.Sprintf("grpc PutAllMetrics OK, %d metrics loaded \n", len(metras))
	models.Sugar.Debug(response.OutData)

	return &response, nil
}

// loadTLSCredentials загрузка сертификатов
func loadTLSCredentials(cert, key string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}
	return credentials.NewTLS(config), nil
}
