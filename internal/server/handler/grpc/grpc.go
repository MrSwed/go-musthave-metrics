package grpc

import (
	"context"
	"errors"
	pb "go-musthave-metrics/internal/grpc/proto"
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/constant"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"
	"go-musthave-metrics/internal/server/service"
	"time"

	"go.uber.org/zap"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	s   *service.Service
	log *zap.Logger
	c   *config.Config
}

func NewMetricsServer(s *service.Service, c *config.Config, log *zap.Logger) *MetricsServer {
	return &MetricsServer{
		s:   s,
		log: log,
		c:   c,
	}
}

func (g *MetricsServer) GetMetric(ctx context.Context, in *pb.GetMetricRequest) (out *pb.GetMetricResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, constant.ServerOperationTimeout*time.Second)
	defer cancel()
	request := in.GetMetric()
	var metric domain.Metric
	metric, err = g.s.GetMetric(ctx, request.Mtype, request.Id)
	if err != nil {
		if errors.Is(err, myErr.ErrNotExist) {
			err = errors.Join(errors.New("metric not exist"), err)
		} else {
			err = errors.Join(errors.New("server error"), err)
			g.log.Error("Error get "+metric.MType, zap.Error(err))
		}
		return
	}
	m := pb.Metric{
		Id:    metric.ID,
		Mtype: metric.MType,
	}
	if metric.Delta != nil {
		m.Delta = int64(*metric.Delta)
	}
	if metric.Value != nil {
		m.Value = float32(*metric.Value)
	}
	out = &pb.GetMetricResponse{
		Metric: &m,
	}
	return
}

func (g *MetricsServer) SetMetric(ctx context.Context, in *pb.SetMetricRequest) (out *pb.SetMetricResponse, err error) {

	return
}

func (g *MetricsServer) SetMetrics(ctx context.Context, in *pb.SetMetricsRequest) (out *pb.SetMetricsResponse, err error) {

	return
}

func (g *MetricsServer) GetMetrics(ctx context.Context, in *pb.GetMetricsRequest) (out *pb.GetMetricsResponse, err error) {

	return
}
