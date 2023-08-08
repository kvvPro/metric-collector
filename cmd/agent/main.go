package main

import (
	"context"

	"github.com/kvvPro/metric-collector/cmd/agent/client"
	"github.com/kvvPro/metric-collector/cmd/agent/config"
	"go.uber.org/zap"
)

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	client.Sugar = *logger.Sugar()

	agentFlags := config.Initialize()
	agent, err := client.NewClient(agentFlags.PollInterval, agentFlags.ReportInterval,
		agentFlags.Address, "text/plain", agentFlags.HashKey, agentFlags.RateLimit)
	if err != nil {
		panic(err)
	}

	client.Sugar.Infow(
		"Starting client",
		"addr", agentFlags.Address,
	)

	agent.Run(context.Background())
}
