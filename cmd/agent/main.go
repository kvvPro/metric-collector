package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/kvvPro/metric-collector/cmd/agent/client"
	"github.com/kvvPro/metric-collector/cmd/agent/config"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	// делаем регистратор SugaredLogger
	client.Sugar = *logger.Sugar()

	agentFlags := initConfigs()
	agent, err := client.NewClient(agentFlags)
	if err != nil {
		client.Sugar.Fatalw(err.Error())
	}

	client.Sugar.Infow(
		"Starting client",
		"addr", agentFlags.Address,
	)

	ctx := context.Background()
	agent.Run(ctx)

	sigQuit := <-shutdown

	agent.Stop()
	client.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
}

func initConfigs() *config.ClientFlags {
	client.Sugar.Infof("\nBuild version: %v", buildVersion)
	client.Sugar.Infof("\nBuild date: %v", buildDate)
	client.Sugar.Infof("\nBuild commit: %v", buildCommit)

	agentFlags, err := config.ReadConfig()
	if err != nil {
		client.Sugar.Fatalw(err.Error(), "event", "read config")
	}
	config.Initialize(agentFlags)
	if err != nil {
		client.Sugar.Fatalw(err.Error(), "event", "read config")
	}
	return agentFlags
}
