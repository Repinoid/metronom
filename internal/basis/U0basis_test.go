package basis

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	
	"gorono/internal/models"
)

type TstBase struct {
	suite.Suite
	t        time.Time
	ctx      context.Context
	dataBase *DBstruct
}

func (suite *TstBase) SetupSuite() { // выполняется перед тестами
	suite.ctx = context.Background()
	suite.t = time.Now()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	models.Sugar = *logger.Sugar()

	log.Println("SetupTest() ---------------------")
}

func (suite *TstBase) TearDownSuite() { // // выполняется после всех тестов
	log.Printf("Spent %v\n", time.Since(suite.t))
}


func TestHandlersSuite(t *testing.T) {
	testBase := new(TstBase)
	testBase.ctx = context.Background()


	log.Println("before run basis.InitDBStorage")
	suite.Run(t, testBase)

}
