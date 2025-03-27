package basis

import (
	"context"
	"gorono/internal/models"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TstBase struct {
	suite.Suite
	t   time.Time
	ctx context.Context
	dataBase    *DBstruct
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

// func (suite *TstHandlers) BeforeTest(suiteName, testName string) { // выполняется перед каждым тестом
// 	var err error
// 	Interbase, err = securitate.ConnectToDB(suite.ctx)
// 	suite.Require().NoErrorf(err, "err %v", err)
// }

// func (suite *TstHandlers) AfterTest(suiteName, testName string) { // // выполняется после каждого теста
//
//		err := Interbase.CloseBase(suite.ctx)
//		suite.Require().NoErrorf(err, "err %v", err)
//	}

func TestHandlersSuite(t *testing.T) {
	testBase := new(TstBase)
	testBase.ctx = context.Background()

	// models.DBEndPoint = "postgres://postgres:passwordas@localhost:5432/forgo"
	// dbStorage, err := InitDBStorage(testBase.ctx, models.DBEndPoint)
	// if err != nil {
	// 	models.Sugar.Debugln("basis.InitDBStorage")
	// 	return
	// }

	// err = dbStorage.TablesDrop(testBase.ctx) // для тестов удаляем таблицы
	// if err != nil {
	// 	models.Sugar.Debugln("table DROP")
	// 	return
	// }
	// dbStorage.DB.Close(testBase.ctx)

	// dbStorage, err = InitDBStorage(testBase.ctx, models.DBEndPoint)
	// if err != nil {
	// 	models.Sugar.Debugln("basis.InitDBStorage 2222")
	// 	return
	// }

	log.Println("before run basis.InitDBStorage")
	suite.Run(t, testBase)

	// testBase.wt = memos.InitMemoryStorage() // тест для базы в памяти
	// log.Println("before run memos.InitMemoryStorage")
	// suite.Run(t, testBase)
}

// go test ./... -v -coverpkg=./...
