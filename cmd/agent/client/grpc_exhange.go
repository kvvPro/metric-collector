package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// импортируем пакет со сгенерированными protobuf-файлами
	"github.com/kvvPro/metric-collector/internal/metrics"
	pb "github.com/kvvPro/metric-collector/proto"
)

func (cli *Client) updateMetrics(ctx context.Context, allMetrics []metrics.Metric) error {
	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа MetricServerClient,
	// через которую будем отправлять сообщения
	c := pb.NewMetricServerClient(conn)

	req := pb.PushMetricsRequest{
		Metric: make([]*pb.Metric, 0),
	}
	for _, el := range allMetrics {
		req.Metric = append(req.Metric, &pb.Metric{
			ID:    el.ID,
			MType: el.MType,
			Delta: *el.Delta,
			Value: *el.Value,
		})
	}

	response, err := c.PushMetrics(ctx, &req)
	if response == nil || err != nil {
		// smth is wrong
		return err
	}

	return nil
}
