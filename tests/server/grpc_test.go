package server_test

import (
	"context"
	"errors"
	"fmt"
	pb "go-musthave-metrics/internal/grpc/proto"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"
	"go-musthave-metrics/internal/server/handler/grpc"
	"math/rand"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func testGRPCGetMetric(suite HandlerTestSuite) {
	t := suite.T()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, "testGauge", testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, "testCounter", testCounter)

	type args struct {
		in *pb.GetMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		wantOut *pb.GetMetricResponse
		wantErr error
	}{
		{
			name: "grpc metric get counter success",
			args: args{

				in: &pb.GetMetricRequest{
					Metric: &pb.Metric{
						Id:    "testCounter",
						Mtype: "counter",
					},
				},
			},
			wantOut: &pb.GetMetricResponse{
				Metric: &pb.Metric{
					Delta: int64(testCounter),
					Id:    "testCounter",
					Mtype: "counter",
				},
			},
			wantErr: nil,
		},

		{
			name: "grpc metric get gauge success",
			args: args{

				in: &pb.GetMetricRequest{
					Metric: &pb.Metric{
						Id:    "testGauge",
						Mtype: "gauge",
					},
				},
			},
			wantOut: &pb.GetMetricResponse{
				Metric: &pb.Metric{
					Value: float32(testGauge),
					Id:    "testGauge",
					Mtype: "gauge",
				},
			},
			wantErr: nil,
		},

		{
			name: "grpc metric get unknown gauge",
			args: args{

				in: &pb.GetMetricRequest{
					Metric: &pb.Metric{
						Id:    "testGaugeUncnownId",
						Mtype: "gauge",
					},
				},
			},

			wantErr: myErr.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := grpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())

			gotOut, err := g.GetMetric(context.Background(), tt.args.in)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("GetMetric() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func testGRPCGetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	ctx := context.Background()
	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)
	_ = suite.Srv().SetGauge(ctx, "testGauge", testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, "testCounter", testCounter)

	type want struct {
		responseContain []string
	}
	tests := []struct {
		name    string
		want    want
		wantErr error
	}{
		{
			name: "grpc metric get html",
			want: want{
				responseContain: []string{"<!doctype html>", "testCounter", "testGauge"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := grpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())

			gotOut, err := g.GetMetrics(context.Background(), nil)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, rc := range tt.want.responseContain {
				assert.Contains(t, string(gotOut.GetHtml()), rc)
			}
		})
	}
}

func testGRPCSetMetric(suite HandlerTestSuite) {
	t := suite.T()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	testCounterPresetName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugePresetName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testCounterName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Intn(500))

	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, testGaugePresetName, testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, testCounterPresetName, testCounter)

	type args struct {
		in *pb.SetMetricRequest
	}
	tests := []struct {
		name    string
		args    args
		wantOut *pb.SetMetricResponse
		wantErr error
	}{
		{
			name: "Save counter. Ok",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Delta: 1,
						Id:    testCounterName,
						Mtype: "counter",
					},
				},
			},
			wantOut: &pb.SetMetricResponse{
				Metric: &pb.Metric{
					Delta: 1,
					Id:    testCounterName,
					Mtype: "counter",
				}},
			wantErr: nil,
		},

		{
			name: "Save gauge. Ok",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Value: 1.1,
						Id:    testGaugeName,
						Mtype: "gauge",
					},
				},
			},
			wantOut: &pb.SetMetricResponse{
				Metric: &pb.Metric{
					Value: 1.1,
					Id:    testGaugeName,
					Mtype: "gauge",
				}},
			wantErr: nil,
		},
		{
			name: "tick counter. Ok",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Delta: 1,
						Id:    testCounterPresetName,
						Mtype: "counter",
					},
				},
			},
			wantOut: &pb.SetMetricResponse{
				Metric: &pb.Metric{
					Delta: 1 + int64(testCounter),
					Id:    testCounterPresetName,
					Mtype: "counter",
				}},
			wantErr: nil,
		},

		{
			name: "update gauge. Ok",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Value: 1.10002,
						Id:    testGaugePresetName,
						Mtype: "gauge",
					},
				},
			},
			wantOut: &pb.SetMetricResponse{
				Metric: &pb.Metric{
					Value: 1.10002,
					Id:    testGaugePresetName,
					Mtype: "gauge",
				}},
			wantErr: nil,
		},
		/* todo: not worked at now * /
		{
			name: "save wrong type",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Delta: 1,
						Id:    testCounterName,
						Mtype: "gauge",
					},
				},
			},
			wantErr: &validator.ValidationErrors{},
		},
		/**/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := grpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())

			gotOut, err := g.SetMetric(context.Background(), tt.args.in)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("SetMetric() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func testGRPCSetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	testCounter := domain.Counter(1)
	testGauge := domain.Gauge(1.0001)

	testCounterPresetName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugePresetName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testCounterName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Intn(500))

	ctx := context.Background()
	_ = suite.Srv().SetGauge(ctx, testGaugePresetName, testGauge)
	_ = suite.Srv().IncreaseCounter(ctx, testCounterPresetName, testCounter)

	type args struct {
		in *pb.SetMetricsRequest
	}
	tests := []struct {
		name    string
		args    args
		wantOut *pb.SetMetricsResponse
		wantErr error
	}{
		{
			name: "Save. Ok",
			args: args{
				in: &pb.SetMetricsRequest{
					Metric: []*pb.Metric{
						{
							Delta: 1,
							Id:    testCounterName,
							Mtype: "counter",
						},
						{
							Id:    testGaugeName,
							Mtype: "gauge",
							Value: 100.0015,
						},
					},
				},
			},
			wantOut: &pb.SetMetricsResponse{
				Metric: []*pb.Metric{
					{
						Delta: 1,
						Id:    testCounterName,
						Mtype: "counter",
					},
					{
						Id:    testGaugeName,
						Mtype: "gauge",
						Value: 100.0015,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Bad metric type",
			args: args{
				in: &pb.SetMetricsRequest{
					Metric: []*pb.Metric{
						{
							Delta: 1,
							Id:    testCounterName,
							Mtype: "unknownType",
						},
						{
							Id:    testGaugeName,
							Mtype: "gauge",
							Value: 100.0015,
						},
					},
				},
			},
			wantErr: &validator.ValidationErrors{},
		},
		/* todo: not worked at now * /
		{
			name: "Wrong types",
			args: args{
				in: &pb.SetMetricsRequest{
					Metric: []*pb.Metric{
						{
							Delta: 1,
							Id:    testCounterName,
							Mtype: "gauge",
						},
						{
							Id:    testGaugeName,
							Mtype: "counter",
							Value: 100.0015,
						},
					},
				},
			},
			wantErr: &validator.ValidationErrors{},
		},
		/**/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := grpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())

			gotOut, err := g.SetMetrics(context.Background(), tt.args.in)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("SetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("SetMetrics() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
