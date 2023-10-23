package main

import (
	"context"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
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

	agentFlags, err := config.ReadConfig()
	if err != nil {
		client.Sugar.Fatalw(err.Error(), "event", "read config")
	}
	config.Initialize(agentFlags)
	if err != nil {
		client.Sugar.Fatalw(err.Error(), "event", "read config")
	}

	agent, err := client.NewClient(agentFlags)
	if err != nil {
		panic(err)
	}

	client.Sugar.Infow(
		"Starting client",
		"addr", agentFlags.Address,
	)

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	asyncCtx, cancelAgent := context.WithCancel(ctx)
	agent.Run(asyncCtx, wg)

	sigQuit := <-shutdown

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
	cancelAgent()
	wg.Wait()
	client.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
}
