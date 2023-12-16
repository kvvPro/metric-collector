package app

import (
	"context"
	"net"
	"time"

	"github.com/kvvPro/metric-collector/internal/metrics"
	ip "github.com/kvvPro/metric-collector/internal/net"
	pb "github.com/kvvPro/metric-collector/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (srv *Server) startGRPCServer(ctx context.Context) {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", srv.Address)
	if err != nil {
		Sugar.Fatal(err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	srv.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(srv.loggingInterceptor,
		srv.validateIPInterceptor))
	// регистрируем сервис
	pb.RegisterMetricServerServer(srv.grpcServer, srv)

	Sugar.Infoln("Сервер gRPC начал работу")
	// получаем запрос gRPC
	go func() {
		if err := srv.grpcServer.Serve(listen); err != nil {
			Sugar.Fatalw(err.Error(), "event", "start grpc server")
		}
	}()
}

func (srv *Server) stopGRPCServer(ctx context.Context) {
	stopped := make(chan struct{})
	Sugar.Infoln("Попытка мягко завершить сервер")
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	go func() {
		defer cancel()
		srv.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-timeout.Done():
		Sugar.Errorf("Ошибка при попытке мягко завершить grpc-сервер: %v", "timeout is expired")
		srv.grpcServer.Stop()
	}
}

func (srv *Server) validateIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var clientIP string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		param := md.Get("X-Real-IP")
		if len(param) > 0 {
			clientIP = param[0]
		}
	}
	if len(clientIP) == 0 {
		return nil, status.Error(codes.Aborted, "not found client IP")
	}
	trusted, err := ip.CheckIPInSubnet(clientIP, srv.TrustedSubnet)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	if !trusted {
		return nil, status.Error(codes.Aborted, "client IP not in trusted subnet")
	}
	return handler(ctx, req)
}

func (srv *Server) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	h, err := handler(ctx, req)

	duration := time.Since(start)

	Sugar.Infoln(
		"uri", info.FullMethod,
		"duration", duration,
		"err", err,
	)
	return h, err
}

func (srv *Server) PushMetrics(ctx context.Context, in *pb.PushMetricsRequest) (*pb.PushMetricsResponse, error) {
	var response pb.PushMetricsResponse

	if err := check(in.Metrics); err != nil {
		return nil, err
	} else {
		var localMetrics = make([]metrics.Metric, 0)
		for _, el := range in.Metrics {
			localMetrics = append(localMetrics, metrics.Metric{
				ID:    el.ID,
				MType: el.MType,
				Delta: el.Delta,
				Value: el.Value,
			})
		}
		err := srv.AddMetricsBatch(ctx, localMetrics)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &response, nil
}

func check(inboundMetrics []*pb.Metric) error {
	for _, m := range inboundMetrics {
		if m.ID == "" {
			return status.Errorf(codes.NotFound, "Missing name of metric")
		}

		if !isValidType(m.MType) {
			return status.Errorf(codes.InvalidArgument, "Invalid type")
		}

		if m.Delta == nil && m.Value == nil ||
			m.MType == metrics.MetricTypeCounter && m.Delta == nil ||
			m.MType == metrics.MetricTypeGauge && m.Value == nil {
			return status.Errorf(codes.InvalidArgument, "Invalid value")
		}
	}

	return nil
}
