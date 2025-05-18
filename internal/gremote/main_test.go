package gremote

// Basic imports
import (
	"context"
	pb "gorono/cmd/proto"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type TSuite struct {
	suite.Suite
	ctx    context.Context
	conn   *grpc.ClientConn
	client pb.MetricClient
}

// !!!!!!!!!!!!!!!!!!!!!!! для тестов ЗАПУСТИТЬ ВРУЧНУЮ СЕРВЕР В cmd/server/

// SetupSuite() запускается перед тестами
func (suite *TSuite) SetupSuite() {
	suite.ctx = context.Background()

	tlsCreds, err := loadClientTLSCredentials("../../cmd/tls/cert.pem")
	suite.Require().NoErrorf(err, "cannot load TLS credentials: %v", err)

	suite.conn, err = grpc.NewClient(":3200", grpc.WithTransportCredentials(tlsCreds))
	suite.Require().NoErrorf(err, "grpc.NewClient %v", err)

	suite.client = pb.NewMetricClient(suite.conn)
	time.Sleep(5 * time.Second)

	log.Println("SetupTest() ---------------------")
}

// SetupSuite() запускается после ВСЕХ тестов
func (suite *TSuite) TearDownSuite() {
	suite.conn.Close()
}

//	func (suite *TSuite) BeforeTest(suiteName, testName string) {
//		log.Println("BeforeTest()", suiteName, testName)
//	}
//
//	func (suite *TSuite) AfterTest(suiteName, testName string) {
//		log.Println("AfterTest()", suiteName, testName)
//	}
func TestGrpcSuite(t *testing.T) {
	log.Println("before run")
	suite.Run(t, new(TSuite))
}
