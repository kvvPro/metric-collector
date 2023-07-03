package main

import (
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
		agentFlags.Host, agentFlags.Port, "text/plain")
	if err != nil {
		panic(err)
	}

	client.Sugar.Infow(
		"Starting client",
		"addr", agentFlags.Host+":"+agentFlags.Port,
	)

	agent.Run()
}
