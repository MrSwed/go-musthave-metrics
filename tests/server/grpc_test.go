package server_test

import (
	"context"
	"errors"
	"fmt"
	pb "go-musthave-metrics/internal/grpc/proto"
	myErr "go-musthave-metrics/internal/server/errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func testGRPCDial(suite HandlerTestSuite, ctx context.Context, meta map[string]string) (ctxOut context.Context, conn *grpc.ClientConn, g pb.MetricsClient, callOpt []grpc.CallOption, err error) {
	callOpt = []grpc.CallOption{}
	ctxOut = ctx
	conn, err = grpc.DialContext(ctx, suite.Cfg().GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return
	}
	if meta != nil {
		metaD := metadata.New(meta)
		ctxOut = metadata.NewOutgoingContext(ctx, metaD)
		callOpt = append(callOpt, grpc.Header(&metaD))
	}
	g = pb.NewMetricsClient(conn)

	return
}

func testGRPCGetMetric(suite HandlerTestSuite) {
	t := suite.T()

	type args struct {
		in *pb.GetMetricRequest
	}
	tests := []struct {
		args    args
		wantOut *pb.GetMetricResponse
		headers map[string]string
		wantErr error
		name    string
	}{
		{
			name: "grpc metric get counter Unauthenticated",
			args: args{
				in: &pb.GetMetricRequest{
					Metric: &pb.Metric{
						Id:    "testCounter-1",
						Mtype: "counter",
					},
				},
			},
			wantErr: status.Error(codes.Unauthenticated, "missing token"),
		},
		{
			name: "grpc metric get counter success",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			args: args{
				in: &pb.GetMetricRequest{
					Metric: &pb.Metric{
						Id:    "testCounter-1",
						Mtype: "counter",
					},
				},
			},
			wantOut: &pb.GetMetricResponse{
				Metric: &pb.Metric{
					Delta: int64(1),
					Id:    "testCounter-1",
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
						Id:    "testGauge-1",
						Mtype: "gauge",
					},
				},
			},
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			wantOut: &pb.GetMetricResponse{
				Metric: &pb.Metric{
					Value: float32(1.0001),
					Id:    "testGauge-1",
					Mtype: "gauge",
				},
			},
			wantErr: nil,
		},

		{
			name: "grpc metric get unknown gauge",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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

			require.NotEmpty(t, tt.args.in.GetMetric()) // more cover proto
			require.NotEmpty(t, tt.args.in.String())    // more cover proto

			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			ctx, conn, g, callOpt, err := testGRPCDial(suite, ctx, tt.headers)
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			gotOut, err := g.GetMetric(ctx, tt.args.in, callOpt...)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantOut != nil && (gotOut == nil || !reflect.DeepEqual(gotOut.GetMetric(), tt.wantOut.GetMetric())) {
				t.Errorf("GetMetric() gotOut.GetMetric() = %v, want.Metric %v", gotOut.GetMetric(), tt.wantOut.GetMetric())
				return
			}

			if tt.wantOut != nil && gotOut != nil { // more cover  protos
				require.Equal(t, gotOut.GetMetric().GetId(), tt.wantOut.GetMetric().GetId())
				require.Equal(t, gotOut.GetMetric().GetValue(), tt.wantOut.GetMetric().GetValue())
				require.Equal(t, gotOut.GetMetric().GetDelta(), tt.wantOut.GetMetric().GetDelta())
				require.Equal(t, gotOut.GetMetric().GetMtype(), tt.wantOut.GetMetric().GetMtype())
				require.Equal(t, gotOut.GetMetric().String(), tt.wantOut.GetMetric().String())
				require.Equal(t, gotOut.String(), tt.wantOut.String())
			}
		})
	}
}

func testGRPCGetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	type want struct {
		responseContain []string
	}

	tests := []struct {
		name    string
		wantErr error
		headers map[string]string
		want    want
	}{
		{
			name:    "try get without token",
			wantErr: status.Error(codes.Unauthenticated, "missing token"),
		},
		{
			name: "grpc metric get html",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			want: want{
				responseContain: []string{"<!doctype html>", "testCounter", "testGauge"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			ctx, conn, g, callOpt, err := testGRPCDial(suite, ctx, tt.headers)
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			gotOut, err := g.GetMetrics(ctx, &pb.GetMetricsRequest{}, callOpt...)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, rc := range tt.want.responseContain {
				assert.Contains(t, string(gotOut.GetHtml()), rc)
				assert.Contains(t, gotOut.String(), rc)
			}
		})
	}
}

func testGRPCSetMetric(suite HandlerTestSuite) {
	t := suite.T()

	testCounterName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Intn(500))

	type args struct {
		in *pb.SetMetricRequest
	}
	tests := []struct {
		args    args
		wantOut *pb.SetMetricResponse
		headers map[string]string
		wantErr error
		name    string
	}{
		{
			name: "Save counter. Unauthenticated",
			args: args{
				in: &pb.SetMetricRequest{
					Metric: &pb.Metric{
						Delta: 1,
						Id:    testCounterName,
						Mtype: "counter",
					},
				},
			},
			wantErr: status.Error(codes.Unauthenticated, "missing token"),
		},

		{
			name: "Save counter. Ok",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
		/* need test data before * /
		{
			name: "tick counter. Ok",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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

			require.NotEmpty(t, tt.args.in.GetMetric()) // more cover proto
			require.NotEmpty(t, tt.args.in.String())    // more cover proto

			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			ctx, conn, g, callOpt, err := testGRPCDial(suite, ctx, tt.headers)
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			gotOut, err := g.SetMetric(ctx, tt.args.in, callOpt...)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantOut != nil && (gotOut == nil || !reflect.DeepEqual(gotOut.GetMetric(), tt.wantOut.GetMetric())) {
				t.Errorf("SetMetric() gotOut.GetMetric() = %v, want.Metric %v", gotOut.GetMetric(), tt.wantOut.GetMetric())
			}
			if tt.wantErr == nil {
				require.Equal(t, gotOut.GetMetric().GetId(), tt.args.in.GetMetric().GetId())
				require.Equal(t, gotOut.GetMetric().GetMtype(), tt.args.in.GetMetric().GetMtype())
				require.Equal(t, gotOut.String(), tt.wantOut.String())
				require.Equal(t, gotOut.GetMetric().String(), tt.wantOut.GetMetric().String())
			}
		})
	}
}

func testGRPCSetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	testCounterName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Intn(500))

	type args struct {
		in *pb.SetMetricsRequest
	}
	tests := []struct {
		args    args
		wantOut *pb.SetMetricsResponse
		headers map[string]string
		wantErr error
		name    string
	}{
		{
			name: "Save. Unauthenticated",
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
			wantErr: status.Error(codes.Unauthenticated, "missing token"),
		},
		{
			name: "Save. Ok",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			wantErr: &validator.ValidationErrors{},
		},
		/* todo: not worked at now * /
		{
			name: "Wrong types",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
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
			require.NotEmpty(t, tt.args.in.GetMetric()) // more cover proto
			require.NotEmpty(t, tt.args.in.String())    // more cover proto

			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			ctx, conn, g, callOpt, err := testGRPCDial(suite, ctx, tt.headers)
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			gotOut, err := g.SetMetrics(ctx, tt.args.in, callOpt...)
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("SetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantOut != nil && (gotOut == nil || !reflect.DeepEqual(gotOut.GetMetric(), tt.wantOut.GetMetric())) {
				t.Errorf("SetMetrics() gotOut.GetMetric() = %v, want.Metric %v", gotOut.GetMetric(), tt.wantOut.GetMetric())
			}
			if tt.wantErr == nil {
				require.Equal(t, gotOut.String(), tt.wantOut.String())
			}
		})
	}
}
