package grpc

import (
	"context"
	"fmt"
	pb "go-musthave-metrics/internal/grpc/proto"
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/service"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Handler struct {
	s   *service.Service
	log *zap.Logger
	c   *config.Config
}

func NewServer(s *service.Service, c *config.Config, log *zap.Logger) *Handler {
	return &Handler{
		s:   s,
		c:   c,
		log: log}
}

func (h *Handler) Handler() (s *grpc.Server) {

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}

	s = grpc.NewServer(grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(h.interceptorLogger(h.log), opts...),
		h.unaryInterceptor,
	))
	pb.RegisterMetricsServer(s, NewMetricsServer(h.s, h.c, h.log))

	return
}

func (h *Handler) unaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if len(h.c.GRPCToken) > 0 {
		var token string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("token"); len(values) > 0 {
				token = values[0]
			}
		}
		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, `missing token`)
		}
		if token != h.c.GRPCToken {
			return nil, status.Error(codes.Unauthenticated, `invalid token`)
		}
	}

	return handler(ctx, req)
}

// interceptorLogger adapts zap logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func (h *Handler) interceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
