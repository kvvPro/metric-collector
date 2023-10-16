package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	rpprof "runtime/pprof"
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
	client.Sugar.Infof("\nBuild version: %v", buildVersion)
	client.Sugar.Infof("\nBuild date: %v", buildDate)
	client.Sugar.Infof("\nBuild commit: %v", buildCommit)

	agentFlags := config.Initialize()
	agent, err := client.NewClient(&agentFlags)
	if err != nil {
		panic(err)
	}

	client.Sugar.Infow(
		"Starting client",
		"addr", agentFlags.Address,
	)

	agent.Run(context.Background())

	sigQuit := <-shutdown
	client.Sugar.Infoln("Server shutdown by signal: ", sigQuit)

	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(agentFlags.MemProfile)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
}
