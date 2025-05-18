package gremote

import (
	"context"
	"fmt"
	_ "net/http/pprof"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "gorono/cmd/proto"

	"gorono/internal/basis"
	"gorono/internal/models"
)

// MetricServer поддерживает все необходимые методы сервера.
type MetricServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName> для совместимости с будущими версиями
	pb.UnimplementedMetricServer
}

// Добавить в базу слайс метрик
func (s *MetricServer) AddBunch(ctx context.Context, in *pb.Bunch) (*pb.BunchResponse, error) {
	var response pb.BunchResponse

	// from []*pb.GMetr to []models.Metrics
	metras := make([]models.Metrics, len(in.Meters))
	for i, v := range in.Meters {
		metras[i].MType = v.GetMType()
		metras[i].ID = v.GetID()
		metras[i].Delta = &v.Delta
		metras[i].Value = &v.Value
	}

	err := basis.RetryMetricWrapper(models.Inter.PutAllMetrics)(ctx, nil, &metras)
	if err != nil {
		models.Sugar.Debugf("grpc PutAllMetrics   err %+v\n", err)
		return nil, status.Errorf(codes.Unimplemented, "grpc PutAllMetrics err %+v code %v\n", err, status.Code(err))
	}
	response.OutData = fmt.Sprintf("grpc PutAllMetrics OK, %d metrics loaded \n", len(metras))
	models.Sugar.Debug(response.OutData)

	return &response, nil
}

// Добавить в базу one metric
func (s *MetricServer) AddOneMetric(ctx context.Context, in *pb.Metr) (*pb.BunchResponse, error) {
	var response pb.BunchResponse

	metr := models.Metrics{}
	metr.MType = in.GetMType()
	metr.ID = in.GetID()
	if metr.MType == "counter" {
		metr.Delta = &in.Delta
	} else {
		metr.Value = &in.Value
	}

	err := basis.RetryMetricWrapper(models.Inter.PutMetric)(ctx, &metr, nil)
	if err != nil {
		models.Sugar.Debugf("grpc PutOneMetric - err %+v\n", err)
		return nil, status.Errorf(codes.Unimplemented, "grpc PutOneMetric err %+v code %v\n", err, status.Code(err))
	}
	response.OutData = fmt.Sprintf("grpc PutOneMetrics OK, %s metric loaded \n", metr.ID)
	models.Sugar.Debug(response.OutData)

	return &response, nil
}

// Добавить в базу one metric
func (s *MetricServer) GetOneMetric(ctx context.Context, in *pb.Metr) (*pb.ReturnOneMetric, error) {
	response := pb.ReturnOneMetric{Meter: &pb.Metr{}}

	metr := models.Metrics{}
	metr.ID = in.GetID()
	if metr.MType = in.GetMType(); metr.MType == "counter" {
		metr.Delta = &in.Delta
	} else {
		metr.Value = &in.Value
	}

	err := basis.RetryMetricWrapper(models.Inter.GetMetric)(ctx, &metr, nil)
	if err != nil {
		models.Sugar.Debugf("grpc GetOneMetric err %+v\n", err)
		return nil, status.Errorf(codes.Unimplemented, "grpc GetOneMetric err %+v \n", err)
	}

	response.Meter.MType = metr.MType
	response.Meter.ID = metr.ID
	if metr.MType == "counter" {
		response.Meter.Delta = *metr.Delta
	} else {
		response.Meter.Value = *metr.Value
	}
	return &response, nil
}

// получить все метрики
func (s *MetricServer) GetAllMetrix(ctx context.Context, in *pb.Bunch) (*pb.ReturnAllMetrics, error) {

	metras := []models.Metrics{}

	err := basis.RetryMetricWrapper(models.Inter.GetAllMetrics)(ctx, nil, &metras)
	if err != nil {
		models.Sugar.Debugf("grpc GetAllMetrics err %+v\n", err)
		return nil, status.Errorf(codes.Unimplemented, "grpc GetAllMetrics err %+v \n", err)
	}

	gMetras := make([]*pb.Metr, len(metras))
	for i, v := range metras {
		if v.MType == "counter" {
			m := pb.Metr{MType: v.MType, ID: v.ID, Delta: *v.Delta}
			gMetras[i] = &m
		} else {
			m := pb.Metr{MType: v.MType, ID: v.ID, Value: *v.Value}
			gMetras[i] = &m
		}
	}

	response := pb.ReturnAllMetrics{Meters: gMetras}

	return &response, nil
}
