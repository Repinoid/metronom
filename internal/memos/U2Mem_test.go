package memos

import (
	"gorono/internal/middlas"
	"gorono/internal/models"
	"os"
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
	err := suite.memorial.PutAllMetrics(suite.ctx, nil, &([]Metrics{m1, m2, m2}))
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

}
