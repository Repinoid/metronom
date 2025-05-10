package basis

import (
	"context"
	"encoding/json"
	"sync"

	"gorono/internal/middlas"
	"gorono/internal/models"
)

func (suite *TstBase) Test00InitDBStorage() {
	tests := []struct {
		name       string
		ctx        context.Context
		dbEndPoint string
		wantErr    bool
	}{
		{
			name:       "InitDB Bad BASE",
			ctx:        context.Background(),
			dbEndPoint: "postgres://postgres:passwordas@localhost:5432/hzwhatbase",
			wantErr:    true,
		},
		{
			name:       "Bad PASSWORD",
			ctx:        context.Background(),
			dbEndPoint: "postgres://postgres:wrongpassword@localhost:5432/forgo",
			wantErr:    true,
		},
		{
			name:       "InitDB Grace manner", // last - RIGHT base params. чтобы база была открыта для дальнейших тестов
			ctx:        context.Background(),
			dbEndPoint: "postgres://postgres:passwordas@localhost:5432/forgo",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			d, err := InitDBStorage(tt.ctx, tt.dbEndPoint)
			suite.dataBase = d
			suite.Require().Equal(err != nil, tt.wantErr) //
			if err == nil {
				suite.dataBase.TablesDrop(tt.ctx)
				//				err = suite.dataBase.DB.Close(tt.ctx)
				suite.Require().NoError(err)
			}
			d, err = InitDBStorage(tt.ctx, tt.dbEndPoint) // reopen after drop
			suite.dataBase = d
			suite.Require().Equal(err != nil, tt.wantErr) //
		})
	}
}

func (suite *TstBase) Test01DBstruct_PutMetric() {

	tests := []struct {
		name    string
		metr    Metrics
		wantErr bool
	}{
		{
			name:    "Nice gauge PUT metric",
			metr:    models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr(777.77)},
			wantErr: false,
		},
		{
			name:    "Nice counter PUT metric",
			metr:    models.Metrics{MType: "counter", ID: "coooo", Delta: middlas.Ptr[int64](777)},
			wantErr: false,
		},
		{
			name:    "Wrong TYPE",
			metr:    models.Metrics{MType: "WTFtype", ID: "nocoooo", Delta: middlas.Ptr[int64](777)},
			wantErr: true,
		},
		{
			name:    "Wrong value instead DELTA",
			metr:    models.Metrics{MType: "counter", ID: "ooo", Value: middlas.Ptr(777.88)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.dataBase.PutMetric(suite.ctx, &tt.metr, nil)
			suite.Require().Equal(err != nil, tt.wantErr)
		})
	}
}
func (suite *TstBase) Test02DBstruct_GetMetric() {

	tests := []struct {
		name          string
		metr, gotmetr Metrics
		wantErr       bool
	}{
		{
			name:    "Nice gauge GET metric",
			metr:    models.Metrics{MType: "gauge", ID: "Alloc"},
			gotmetr: models.Metrics{MType: "gauge", ID: "Alloc", Value: middlas.Ptr(777.77)},
			wantErr: false,
		},
		{
			name:    "Nice counter GET metric",
			metr:    models.Metrics{MType: "counter", ID: "coooo"},
			gotmetr: models.Metrics{MType: "counter", ID: "coooo", Delta: middlas.Ptr[int64](777)},
			wantErr: false,
		},
		{
			name:    "bad GET, no ID",
			metr:    models.Metrics{MType: "counter", ID: "c"},
			gotmetr: models.Metrics{MType: "counter", ID: "c", Delta: middlas.Ptr[int64](777)},
			wantErr: true,
		},
		{
			name:    "bad Type",
			metr:    models.Metrics{MType: "count", ID: "coooo"},
			gotmetr: models.Metrics{MType: "counter", ID: "coooo", Delta: middlas.Ptr[int64](777)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.dataBase.GetMetric(suite.ctx, &tt.metr, nil)
			suite.Require().Equal(err != nil, tt.wantErr)
			if err == nil {
				gm, _ := json.Marshal(tt.gotmetr)
				gm1, _ := json.Marshal(tt.metr)

				//	reflect.DeepEqual(m1, m2)
				suite.Require().JSONEq(string(gm), string(gm1))
			}
		})
	}
}

func (suite *TstBase) Test02DBstruct_PutAllMetrics() {

	tests := []struct {
		name    string
		metras  []Metrics
		wantErr bool
	}{
		{
			name: "Nice gauge PUT all metrics",
			//metr:    models.Metrics{MType: "gauge", ID: "Alloc"},
			metras: []models.Metrics{{MType: "gauge", ID: "Ga", Value: middlas.Ptr(777.77)},
				{MType: "counter", ID: "Co", Delta: middlas.Ptr[int64](777)}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.dataBase.PutAllMetrics(suite.ctx, nil, &tt.metras) //*[]Metrics)
			// suite.dataBase.GetMetric(suite.ctx, &tt.metr, nil)
			suite.Require().Equal(err != nil, tt.wantErr)
		})
	}
}
func (suite *TstBase) Test02XDBstruct_GetAllMetrics() {

	tests := []struct {
		name         string
		metras, outm []Metrics
		wantErr      bool
	}{
		{
			name: "Nice gauge GET all metrics",
			//metr:    models.Metrics{MType: "gauge", ID: "Alloc"},
			metras: []models.Metrics{{MType: "gauge", ID: "Ga"},
				{MType: "counter", ID: "Co", Delta: middlas.Ptr[int64](777)}},
			outm: []models.Metrics{{MType: "gauge", ID: "Ga"},
				{MType: "counter", ID: "Co", Delta: middlas.Ptr[int64](777)}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.dataBase.GetAllMetrics(suite.ctx, nil, &tt.metras) //*[]Metrics)
			// suite.dataBase.GetMetric(suite.ctx, &tt.metr, nil)
			suite.Require().Equal(err != nil, tt.wantErr)
			// if err == nil {
			// 	gm, _ := json.Marshal(tt.metras)
			// 	gm1, _ := json.Marshal(tt.outm)

			// 	suite.Require().JSONEq(string(gm), string(gm1))
			// }
		})
	}
}
func (suite *TstBase) Test02XXDBstruct_rests() {
	err := suite.dataBase.LoadMS("s")
	suite.Require().NoError(err)
	err = suite.dataBase.SaveMS("s")
	suite.Require().NoError(err)

	var wg sync.WaitGroup
	wg.Add(1)
	err = suite.dataBase.Saver(suite.ctx, "", 0, &wg)

	suite.Require().NoError(err)
	err = suite.dataBase.Ping(suite.ctx, "")
	suite.Require().NoError(err)
	fio := suite.dataBase.GetName()
	suite.Require().Equal(fio, "DBaser")
}
func (suite *TstBase) Test02XXX_Close() {
	suite.dataBase.Close()
}
func (suite *TstBase) TestXX_PingAfterDBClose() {
	err := suite.dataBase.Ping(suite.ctx, "")
	// ERROR
	suite.Require().Error(err)

	metras := []models.Metrics{{MType: "gauge", ID: "Ga"},
		{MType: "counter", ID: "Co", Delta: middlas.Ptr[int64](777)}}
	err = suite.dataBase.PutAllMetrics(suite.ctx, nil, &metras)
	suite.Require().Error(err)
	err = suite.dataBase.GetAllMetrics(suite.ctx, nil, &metras)
	suite.Require().Error(err)
	err = suite.dataBase.PutMetric(suite.ctx, &metras[0], nil)
	suite.Require().Error(err)
	err = suite.dataBase.GetMetric(suite.ctx, &metras[0], nil)
	suite.Require().Error(err)
}
