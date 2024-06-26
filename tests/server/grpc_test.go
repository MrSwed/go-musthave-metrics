package server_test

import (
	"context"
	"errors"
	"fmt"
	pb "go-musthave-metrics/internal/grpc/proto"
	"go-musthave-metrics/internal/server/domain"
	myErr "go-musthave-metrics/internal/server/errors"
	myGrpc "go-musthave-metrics/internal/server/handler/grpc"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func testGRPCDial(suite HandlerTestSuite, ctx context.Context, meta map[string]string) (ctxOut context.Context, conn *grpc.ClientConn, grpcClient pb.MetricsClient, callOpt []grpc.CallOption, err error) {
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
	grpcClient = pb.NewMetricsClient(conn)

	return
}

func testGRPCProto(suite HandlerTestSuite) {
	t := suite.T()
	t.Run("generated grpc proto not implemented", func(t *testing.T) {
		g := myGrpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())
		ctx := context.Background()
		_, err := g.UnimplementedMetricsServer.GetMetrics(ctx, nil)
		assert.Error(t, err, status.Errorf(codes.Unimplemented, "method GetMetrics not implemented"))
		_, err = g.UnimplementedMetricsServer.GetMetric(ctx, nil)
		assert.Error(t, err, status.Errorf(codes.Unimplemented, "method GetMetric not implemented"))
		_, err = g.UnimplementedMetricsServer.SetMetrics(ctx, nil)
		assert.Error(t, err, status.Errorf(codes.Unimplemented, "method SetMetrics not implemented"))
		_, err = g.UnimplementedMetricsServer.SetMetric(ctx, nil)
		assert.Error(t, err, status.Errorf(codes.Unimplemented, "method SetMetric not implemented"))
	})
}

func testGRPCGetMetric(suite HandlerTestSuite) {
	t := suite.T()

	type args struct {
		in         *pb.GetMetricRequest
		useSrvServ bool
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
			name: "grpc metric get counter from service success",
			args: args{
				useSrvServ: true,
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

			require.NotEmpty(t, tt.args.in.GetMetric())
			require.NotEmpty(t, tt.args.in.String())
			db, di := tt.args.in.Descriptor()
			require.NotEmpty(t, db)
			require.NotEmpty(t, di)

			ctx := context.Background()
			var (
				gotOut   *pb.GetMetricResponse
				err      error
				conn     *grpc.ClientConn
				pbClient pb.MetricsClient
				callOpt  []grpc.CallOption
			)
			if tt.args.useSrvServ {
				g := myGrpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())
				gotOut, err = g.GetMetric(ctx, tt.args.in)
			} else {
				ctx, conn, pbClient, callOpt, err = testGRPCDial(suite, ctx, tt.headers)
				require.NoError(t, err)
				defer func() { require.NoError(t, conn.Close()) }()
				gotOut, err = pbClient.GetMetric(ctx, tt.args.in, callOpt...)
			}
			if (err != nil) != (tt.wantErr != nil) ||
				(tt.wantErr != nil && !errors.Is(err, tt.wantErr) && !assert.Contains(t, err.Error(), tt.wantErr.Error())) {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantOut != nil && (gotOut == nil || !reflect.DeepEqual(gotOut.GetMetric(), tt.wantOut.GetMetric())) {
				t.Errorf("GetMetric() gotOut.GetMetric() = %v, want.Metric %v", gotOut.GetMetric(), tt.wantOut.GetMetric())
				return
			}

			if tt.wantOut != nil && gotOut != nil {
				require.Equal(t, gotOut.GetMetric().GetId(), tt.wantOut.GetMetric().GetId())
				require.Equal(t, gotOut.GetMetric().GetValue(), tt.wantOut.GetMetric().GetValue())
				require.Equal(t, gotOut.GetMetric().GetDelta(), tt.wantOut.GetMetric().GetDelta())
				require.Equal(t, gotOut.GetMetric().GetMtype(), tt.wantOut.GetMetric().GetMtype())
				require.Equal(t, gotOut.GetMetric().String(), tt.wantOut.GetMetric().String())
				require.Equal(t, gotOut.String(), tt.wantOut.String())
				db, di = gotOut.GetMetric().Descriptor()
				require.NotEmpty(t, db)
				require.NotEmpty(t, di)
				db, di = gotOut.Descriptor()
				require.NotEmpty(t, db)
				require.NotEmpty(t, di)
			}
		})
	}
}

func testGRPCGetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	type args struct {
		useSrvServ bool
	}
	type want struct {
		responseContain []string
	}

	tests := []struct {
		headers map[string]string
		name    string
		wantErr error
		want    want
		args    args
	}{
		{
			name:    "try get without token",
			wantErr: status.Error(codes.Unauthenticated, "missing token"),
		},
		{
			name: "try get wit invalid token",
			headers: map[string]string{
				"token": "invalid token",
			},
			wantErr: status.Error(codes.Unauthenticated, "invalid token"),
		},
		{
			name: "grpc app metric get html",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			want: want{
				responseContain: []string{"<!doctype html>", "testCounter", "testGauge"},
			},
			wantErr: nil,
		},
		{
			name: "grpc srv metric get html",
			args: args{
				useSrvServ: true,
			},
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

			var (
				gotOut   *pb.GetMetricsResponse
				err      error
				conn     *grpc.ClientConn
				pbClient pb.MetricsClient
				callOpt  []grpc.CallOption
			)
			if tt.args.useSrvServ {
				g := myGrpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())
				gotOut, err = g.GetMetrics(ctx, nil)
			} else {
				ctx, conn, pbClient, callOpt, err = testGRPCDial(suite, ctx, tt.headers)
				require.NoError(t, err)
				defer func() { require.NoError(t, conn.Close()) }()
				gotOut, err = pbClient.GetMetrics(ctx, &pb.GetMetricsRequest{}, callOpt...)
			}
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
		in         *pb.SetMetricRequest
		useSrvServ bool
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
		{
			name: "tick counter. Ok",
			headers: map[string]string{
				"token": suite.Cfg().GRPCToken,
			},
			args: args{
				useSrvServ: true,
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

			require.NotEmpty(t, tt.args.in.GetMetric())
			require.NotEmpty(t, tt.args.in.String())
			db, di := tt.args.in.Descriptor()
			require.NotEmpty(t, db)
			require.NotEmpty(t, di)

			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()

			var (
				gotOut   *pb.SetMetricResponse
				err      error
				conn     *grpc.ClientConn
				pbClient pb.MetricsClient
				callOpt  []grpc.CallOption
			)
			if tt.args.useSrvServ {
				g := myGrpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())
				gotOut, err = g.SetMetric(ctx, tt.args.in)
			} else {
				ctx, conn, pbClient, callOpt, err = testGRPCDial(suite, ctx, tt.headers)
				require.NoError(t, err)
				defer func() { require.NoError(t, conn.Close()) }()
				gotOut, err = pbClient.SetMetric(ctx, tt.args.in, callOpt...)
			}

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
				db, di := gotOut.Descriptor()
				require.NotEmpty(t, db)
				require.NotEmpty(t, di)
			}
		})
	}
}

func testGRPCSetMetrics(suite HandlerTestSuite) {
	t := suite.T()

	testCounterName := fmt.Sprintf("testCounter%d", rand.Intn(500))
	testGaugeName := fmt.Sprintf("testGauge%d", rand.Intn(500))

	type args struct {
		in         *pb.SetMetricsRequest
		useSrvServ bool
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
			require.NotEmpty(t, tt.args.in.GetMetric())
			require.NotEmpty(t, tt.args.in.String())
			db, di := tt.args.in.Descriptor()
			require.NotEmpty(t, db)
			require.NotEmpty(t, di)

			ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()

			var (
				gotOut   *pb.SetMetricsResponse
				err      error
				conn     *grpc.ClientConn
				pbClient pb.MetricsClient
				callOpt  []grpc.CallOption
			)
			if tt.args.useSrvServ {
				g := myGrpc.NewMetricsServer(suite.Srv(), suite.Cfg(), zap.NewNop())
				gotOut, err = g.SetMetrics(ctx, tt.args.in)
			} else {
				ctx, conn, pbClient, callOpt, err = testGRPCDial(suite, ctx, tt.headers)
				require.NoError(t, err)
				defer func() { require.NoError(t, conn.Close()) }()
				gotOut, err = pbClient.SetMetrics(ctx, tt.args.in, callOpt...)
			}

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
				db, di := gotOut.Descriptor()
				require.NotEmpty(t, db)
				require.NotEmpty(t, di)
			}
		})
	}
}
