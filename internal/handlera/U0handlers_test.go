package handlera

import (
	"context"
	"gorono/internal/basis"
	"gorono/internal/memos"
	"gorono/internal/models"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TstHandlers struct {
	suite.Suite
	//	cmnd *exec.Cmd
	t   time.Time
	ctx context.Context
	wt  models.Interferon
}

func (suite *TstHandlers) SetupSuite() { // выполняется перед тестами
	//var err error
	suite.ctx = context.Background()
	suite.t = time.Now()

	//memStor := memos.InitMemoryStorage()
	models.Inter = suite.wt

	// dbStorage, err := basis.InitDBStorage(suite.ctx, models.DBEndPoint)
	// suite.Require().NoErrorf(err, "err %v", err)
	// for _, tab := range []string{"orders", "tokens", "withdrawn", "accounts"} {
	// 	dropOrder := "DROP TABLE " + tab + " ;"
	// 	_, err := dbStorage.DB.Exec(suite.ctx, dropOrder)
	// 	suite.Require().NoErrorf(err, "err %v", err)
	// }
	// dbStorage.DB.Close(suite.ctx)

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

	testHandler.wt = dbStorage // тест для базы в постгрес
	log.Println("before run basis.InitDBStorage")
	suite.Run(t, testHandler)

	testHandler.wt = memos.InitMemoryStorage() // тест для базы в памяти
	log.Println("before run memos.InitMemoryStorage")
	suite.Run(t, testHandler)
}

// go test ./... -v -coverpkg=./...
