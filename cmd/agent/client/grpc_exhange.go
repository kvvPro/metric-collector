package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	// импортируем пакет со сгенерированными protobuf-файлами
	"github.com/kvvPro/metric-collector/internal/metrics"
	ip "github.com/kvvPro/metric-collector/internal/net"
	pb "github.com/kvvPro/metric-collector/proto"
)

func (cli *Client) updateMetrics(ctx context.Context, allMetrics []metrics.Metric) error {
	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(cli.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа MetricServerClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricServerClient(conn)

	req := pb.PushMetricsRequest{
		Metrics: make([]*pb.Metric, 0),
	}
	for _, el := range allMetrics {
		req.Metrics = append(req.Metrics, &pb.Metric{
			ID:    el.ID,
			MType: el.MType,
			Delta: el.Delta,
			Value: el.Value,
		})
	}

	localIP := ip.GetOutboundIP(cli.Address)
	md := metadata.New(map[string]string{"X-Real-IP": localIP.String()})
	ctxClient := metadata.NewOutgoingContext(ctx, md)

	response, err := c.PushMetrics(ctxClient, &req)
	if response == nil || err != nil {
		// smth is wrong
		return err
	}

	return nil
}
