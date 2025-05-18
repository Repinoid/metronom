package gremote

// Basic imports
import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gorono/internal/memos"
	"gorono/internal/models"
	"log"
	"os"

	pb "gorono/cmd/proto"

	"google.golang.org/grpc/credentials"
)

func (suite *TSuite) Test01PutBunch() {

	//memStorage := []models.Metrics{}
	memStorage := *memos.GetMetrixFromOS()
	addMetrix := *memos.GetMoreMetrix()
	memStorage = append(memStorage, addMetrix...)
	err := sendTestMetrics(suite.ctx, memStorage, suite)
	suite.Require().NoError(err)

	log.Println("testexample5")

}

func (suite *TSuite) Test02PutOneMetric() {
	d := int64(314156)
	v := models.Metrics{MType: "counter", ID: "HZ", Delta: &d}
	gMetr := pb.Metr{MType: v.MType, ID: v.ID, Delta: *v.Delta}

	resp, err := suite.client.AddOneMetric(suite.ctx, &gMetr)
	suite.Require().NoError(err)

	if resp.Error != "" {
		fmt.Println(resp.Error)
	}
	fmt.Printf("Client %s\n", resp.OutData)
}
func (suite *TSuite) Test03GetOneMetric() {
	// HZ exists
	v := models.Metrics{MType: "counter", ID: "HZ"}
	gMetr := pb.Metr{MType: v.MType, ID: v.ID}
	resp, err := suite.client.GetOneMetric(suite.ctx, &gMetr)
	suite.Require().NoError(err)
	suite.Require().GreaterOrEqual(resp.Meter.Delta, int64(314159))

	// HZ1 not exists
	v1 := models.Metrics{MType: "counter", ID: "HZ1"}
	gMetr1 := pb.Metr{MType: v1.MType, ID: v1.ID}
	_, err = suite.client.GetOneMetric(suite.ctx, &gMetr1)
	suite.Require().Error(err)
}
func (suite *TSuite) Test04GetAllMetrics() {
	gMetras := pb.Bunch{Meters: []*pb.Metr{}}
	resp, err := suite.client.GetAllMetrix(suite.ctx, &gMetras)
	suite.Require().NoError(err)

	m := resp.Meters
	for _, v := range m {
		log.Printf("%+v\n", v)
	}
}

func sendTestMetrics(ctx context.Context, bunch []models.Metrics, suite *TSuite) (err error) {

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
	// var conn *grpc.ClientConn
	// tlsCreds, err := loadClientTLSCredentials("../../cmd/tls/cert.pem")
	// if err != nil {
	// 	log.Printf("cannot load TLS credentials: %v", err)
	// 	return
	// }
	// conn, err = grpc.NewClient(":3200", grpc.WithTransportCredentials(tlsCreds))

	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// defer conn.Close()

	// client := pb.NewMetricClient(conn)

	resp, err := suite.client.AddBunch(ctx, &pb.Bunch{
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
