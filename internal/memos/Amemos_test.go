package memos

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"gorono/internal/models"
)

type TstMemo struct {
	suite.Suite
	startTime time.Time
	ctx       context.Context
	memorial  *MemoryStorageStruct
}

// выполняется перед тестами
func (suite *TstMemo) SetupSuite() {
	suite.ctx = context.Background()
	suite.startTime = time.Now()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	models.Sugar = *logger.Sugar()

	suite.memorial = InitMemoryStorage()

	log.Println("SetupTest() ---------------------")
}

// выполняется после всех тестов
func (suite *TstMemo) TearDownSuite() { //
	log.Printf("Spent %v\n", time.Since(suite.startTime))
}


func TestHandlersSuite(t *testing.T) {
	testMemos := new(TstMemo)
	testMemos.ctx = context.Background()

	log.Println("before run tests")
	suite.Run(t, testMemos)

}
