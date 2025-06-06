package memos

import (
	"context"
	"os"
	"sync"

	"gorono/internal/middlas"
	"gorono/internal/models"
)

func (suite *TstMemo) Test02MemPutGetMetric() {

	tests := []struct {
		name       string
		metr       models.Metrics
		marchalled []byte
		wantErr    bool
	}{
		{
			name:    "Nice gauge",
			metr:    models.Metrics{MType: "gauge", ID: "gaaga", Value: middlas.Ptr(777.77)},
			wantErr: false,
		},
		{
			name:    "Bad gauge",
			metr:    models.Metrics{MType: "gauge", ID: "gaaga1", Delta: middlas.Ptr[int64](777)},
			wantErr: true,
		},
		{
			name:    "Nice counter",
			metr:    models.Metrics{MType: "counter", ID: "cunt", Delta: middlas.Ptr[int64](777)},
			wantErr: false,
		},
		{
			name:    "Bad counter",
			metr:    models.Metrics{MType: "counter", ID: "kunka", Value: middlas.Ptr(777.77)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.memorial.PutMetric(suite.ctx, &tt.metr, nil)
			suite.Require().Equal(tt.wantErr, err != nil, err)
			if err == nil {

				metr := models.Metrics{MType: tt.metr.MType, ID: tt.metr.ID}
				err = suite.memorial.GetMetric(suite.ctx, &metr, nil)
				suite.Require().NoError(err)
				suite.Require().EqualValues(metr.Value, tt.metr.Value) // EqualValues сравнивает значения и по ссылкам. и nil
				suite.Require().EqualValues(metr.Delta, tt.metr.Delta)

			}
		})
	}

}
func (suite *TstMemo) Test03Mem() {
	err := suite.memorial.PutAllMetrics(suite.ctx, nil, &([]models.Metrics{m1, m2, m2}))
	suite.Require().NoError(err)
	metr := models.Metrics{MType: m1.MType, ID: m1.ID}
	err = suite.memorial.GetMetric(suite.ctx, &metr, nil)
	suite.Require().NoError(err)
	suite.Require().EqualValues(m1.Value, metr.Value)
	suite.Require().EqualValues(m1.Delta, metr.Delta)

	metras := []models.Metrics{}
	err = suite.memorial.GetAllMetrics(suite.ctx, nil, &metras)
	suite.Require().NoError(err)
	suite.Require().Equal(len(metras), 4)
}

func (suite *TstMemo) Test04Mem() {
	med, err := suite.memorial.MarshalMS()
	suite.Require().NoError(err)

	tstMemorial := InitMemoryStorage()
	err = tstMemorial.UnmarshalMS(med)
	suite.Require().NoError(err)

	suite.Require().EqualValues(suite.memorial, tstMemorial)

	suite.Require().Equal("Memorial", suite.memorial.GetName())

	err = suite.memorial.Ping(suite.ctx, "postgres://postgres:passwordas@localhost:5432/forgo")
	suite.Require().NoError(err)
	err = suite.memorial.Ping(suite.ctx, "postgres://postgres:wrongpassword@localhost:5432/forgo")
	suite.Require().Error(err)

	err = suite.memorial.SaveMS("kut.metr")
	suite.Require().NoError(err)

	err = tstMemorial.LoadMS("kut.metr")
	suite.Require().NoError(err)
	suite.Require().EqualValues(suite.memorial, tstMemorial)

	err = os.Remove("kut.metr")
	suite.Require().NoError(err)

	metras := GetMetrixFromOS()
	suite.Require().Equal(len(*metras), 29)

	metras = GetMoreMetrix()
	suite.Require().Equal(len(*metras), 3)

	suite.memorial.Close()

}

func (suite *TstMemo) Test05Saver() {

	var wg sync.WaitGroup
	wg.Add(1)
	err := suite.memorial.Saver(suite.ctx, "tt://f.out", 1, &wg)
	suite.Require().Error(err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	wg.Add(1)
	err = suite.memorial.Saver(ctx, "f.out", 1, &wg)
	suite.Require().Contains(err.Error(), "Saver остановлен по сигналу")

}
