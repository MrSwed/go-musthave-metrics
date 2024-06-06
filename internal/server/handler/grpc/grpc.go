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

	"github.com/go-playground/validator/v10"
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

func pbSetFromMetric(metrics ...domain.Metric) (m []*pb.Metric) {
	m = make([]*pb.Metric, len(metrics))
	for i := 0; i < len(metrics); i++ {
		m[i] = &pb.Metric{
			Id:    metrics[i].ID,
			Mtype: metrics[i].MType,
		}
		if metrics[i].Delta != nil {
			m[i].Delta = int64(*metrics[i].Delta)
		}
		if metrics[i].Value != nil {
			m[i].Value = float32(*metrics[i].Value)
		}
	}
	return
}

func metricSetFromPb(metrics ...*pb.Metric) (m []domain.Metric) {
	m = make([]domain.Metric, len(metrics))
	for i := 0; i < len(metrics); i++ {
		m[i] = domain.Metric{
			Delta: &[]domain.Counter{domain.Counter(metrics[i].Delta)}[0],
			Value: &[]domain.Gauge{domain.Gauge(metrics[i].Value)}[0],
			ID:    metrics[i].Id,
			MType: metrics[i].Mtype,
		}
	}
	return
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

	out = &pb.GetMetricResponse{
		Metric: pbSetFromMetric(metric)[0],
	}
	return
}

func (g *MetricsServer) SetMetric(ctx context.Context, in *pb.SetMetricRequest) (out *pb.SetMetricResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, constant.ServerOperationTimeout*time.Second)
	defer cancel()
	request := in.GetMetric()
	var metric domain.Metric
	metrikIn := metricSetFromPb(request)[0]
	if metric, err = g.s.SetMetric(ctx, metrikIn); err != nil {
		if errors.As(err, &validator.ValidationErrors{}) {
			err = errors.Join(errors.New("bad input data: "), err)
		} else {
			err = errors.Join(errors.New("error set metric: "), err)
			g.log.Error("Error set metric", zap.Error(err))
		}
		return
	}

	out = &pb.SetMetricResponse{
		Metric: pbSetFromMetric(metric)[0],
	}
	return
}

func (g *MetricsServer) SetMetrics(ctx context.Context, in *pb.SetMetricsRequest) (out *pb.SetMetricsResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, constant.ServerOperationTimeout*time.Second)
	defer cancel()
	request := in.GetMetric()
	var metrics []domain.Metric
	metricsIn := metricSetFromPb(request...)
	if metrics, err = g.s.SetMetrics(ctx, metricsIn); err != nil {
		if errors.As(err, &validator.ValidationErrors{}) {
			err = errors.Join(errors.New("bad input data: "), err)
		} else {
			err = errors.Join(errors.New("error set metrics: "), err)
			g.log.Error("Error set metrics", zap.Error(err))
		}
		return
	}

	out = &pb.SetMetricsResponse{
		Metric: pbSetFromMetric(metrics...),
	}

	return
}

func (g *MetricsServer) GetMetrics(ctx context.Context, _ *pb.GetMetricsRequest) (out *pb.GetMetricsResponse, err error) {
	ctx, cancel := context.WithTimeout(ctx, constant.ServerOperationTimeout*time.Second)
	defer cancel()
	var html []byte
	html, err = g.s.GetMetricsHTMLPage(ctx)
	if err != nil {
		if errors.Is(err, myErr.ErrNotExist) {
			err = errors.Join(errors.New("metrics not exist"), err)
		} else {
			err = errors.Join(errors.New("server error"), err)
			g.log.Error("Error get metrics", zap.Error(err))
		}
		return
	}

	out = &pb.GetMetricsResponse{
		Html: html,
	}

	return
}
