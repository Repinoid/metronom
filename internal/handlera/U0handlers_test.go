package handlera

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
)

type TstHandlers struct {
	suite.Suite
	//	cmnd *exec.Cmd
	t   time.Time
	ctx context.Context
	wt  models.Interferon
}

// некоторые закоулки if err != nil просто недостижимы, т.к. корректность метрик проверяется неоднократно. поэтому coverage недостаточно высок

func (suite *TstHandlers) SetupSuite() { // выполняется перед тестами
	suite.ctx = context.Background()
	suite.t = time.Now()

	models.Inter = suite.wt

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	models.Sugar = *logger.Sugar()

	log.Println("SetupTest() ---------------------")
}

func (suite *TstHandlers) TearDownSuite() { // // выполняется после всех тестов
	log.Printf("Spent %v\n", time.Since(suite.t))
}

func TestHandlersSuite(t *testing.T) {
	testHandler := new(TstHandlers)
	testHandler.ctx = context.Background()

	models.DBEndPoint = "postgres://postgres:passwordas@localhost:5432/forgo"
	dbStorage, err := basis.InitDBStorage(testHandler.ctx, models.DBEndPoint)
	if err != nil {
		log.Println("basis.InitDBStorage")
		return
	}

	err = dbStorage.TablesDrop(testHandler.ctx) // для тестов удаляем таблицы
	if err != nil {
		log.Println("table DROP")
		return
	}
	dbStorage.DB.Close()

	dbStorage, err = basis.InitDBStorage(testHandler.ctx, models.DBEndPoint)
	if err != nil {
		log.Println("basis.InitDBStorage 2222")
		return
	}
	// тест для базы в постгрес
	testHandler.wt = dbStorage
	log.Println("before run basis.InitDBStorage")
	suite.Run(t, testHandler)

	// тест для базы в памяти
	testHandler.wt = memos.InitMemoryStorage()
	log.Println("before run memos.InitMemoryStorage")
	suite.Run(t, testHandler)
}
