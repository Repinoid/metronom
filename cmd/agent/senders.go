package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	pb "gorono/cmd/proto"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"
)

// sendMetricsByHttp посылает слайс метрик на сервер by HTTP
func sendMetricsByHTTP(ctx context.Context, bunch []models.Metrics) (err error) {

	marshalledBunch, err := json.Marshal(bunch)
	if err != nil {
		log.Println(err)
		return
	}
	if cryptoKeyFile != "" {

		coded, err := privacy.Encrypt(marshalledBunch, cryptoKey)

		if err != nil {
			log.Println(err)
			return err
		}
		marshalledBunch = coded
	}
	compressedBunch, err := middlas.Pack2gzip(marshalledBunch)

	if err != nil {
		log.Println(err)
		return
	}

	httpc := resty.New() //
	httpc.SetBaseURL("http://" + host)

	httpc.SetRetryCount(3)
	httpc.SetRetryWaitTime(1 * time.Second)    // начальное время повтора
	httpc.SetRetryMaxWaitTime(9 * time.Second) // 1+3+5
	httpc.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		select {
		case <-ctx.Done():
			httpc.SetRetryCount(0)
			//  RetryAfterFunc Non-nil error is returned if it is found that the request is not retryable
			return 0, errors.New("stop to retry")
		default:
			rwt := client.RetryWaitTime
			client.SetRetryWaitTime(rwt + 2*time.Second) //	увеличение времени ожидания на 2 сек
			return client.RetryWaitTime, nil
		}
	})

	req := httpc.R().
		SetHeader("Content-Encoding", "gzip"). // сжаtо
		SetBody(compressedBunch).
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", GetLocalIP()).
		SetDoNotParseResponse(false)

	resp, _ := req.
		Post("/updates/") // slash on the tile

	if resp.StatusCode() == http.StatusOK { // при успешной отправке метрик обнуляем cчётчик
		atomic.StoreInt64(&cunt, 0) //	cunt = 0
	}
	log.Printf("AGENT responce from server %+v\n", resp.StatusCode())

	return nil
}

func loadClientTLSCredentials(cert string) (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(cert)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}
	config := &tls.Config{
		// Set InsecureSkipVerify to skip the default validation we are
		// replacing. This will not disable VerifyConnection.
		InsecureSkipVerify: true,
		RootCAs:            certPool,
	}
	return credentials.NewTLS(config), nil
}

func sendMetricsByGrpc(ctx context.Context, bunch []models.Metrics) (err error) {

	// from  []models.Metrics to []*pb.GMetr
	gMetras := make([]*pb.Metr, len(bunch))
	//gMetras := []*pb.Metr{}
	for i, v := range bunch {
		if v.MType == "counter" {
			m := pb.Metr{MType: v.MType, ID: v.ID, Delta: *v.Delta}
			gMetras[i] = &m
		} else {
			m := pb.Metr{MType: v.MType, ID: v.ID, Value: *v.Value}
			gMetras[i] = &m
		}
	}

	var conn *grpc.ClientConn

	// устанавливаем соединение с сервером
	tlsCreds, err := loadClientTLSCredentials(cryptoKeyFile)
	if err != nil {
		log.Printf("cannot load TLS credentials: %v", err)
		return
	}
	conn, err = grpc.NewClient(gPort, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	client := pb.NewMetricClient(conn)

	md := metadata.New(map[string]string{"X-Real-IP": GetLocalIP()})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.AddBunch(ctx, &pb.Bunch{
		Meters: gMetras,
	})
	if err != nil {
		log.Println(err)
		return
	}
	if resp.Error != "" {
		fmt.Println(resp.Error)
	}
	fmt.Printf("Client %s\n", resp.OutData)

	return
}
