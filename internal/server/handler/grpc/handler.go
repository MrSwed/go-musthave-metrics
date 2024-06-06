package grpc

import (
	"context"
	pb "go-musthave-metrics/internal/grpc/proto"
	"go-musthave-metrics/internal/server/config"
	"go-musthave-metrics/internal/server/service"

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
	s = grpc.NewServer(grpc.UnaryInterceptor(h.unaryInterceptor))
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
