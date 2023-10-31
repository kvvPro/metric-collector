package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	app "github.com/kvvPro/metric-collector/cmd/server/app"
	"github.com/kvvPro/metric-collector/cmd/server/config"

	"go.uber.org/zap"
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
	app.Sugar = *logger.Sugar()

	srvFlags := initConfigs()
	app.Sugar.Infoln("after init config")

	srv, err := app.NewServer(srvFlags)

	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "create server")
	}

	ctx := context.Background()
	srv.StartServer(ctx, srvFlags)

	sigQuit := <-shutdown

	srv.StopServer(ctx)
	app.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
}

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func initConfigs() *config.ServerFlags {

	app.Sugar.Infof("\nBuild version: %v", buildVersion)
	app.Sugar.Infof("\nBuild date: %v", buildDate)
	app.Sugar.Infof("\nBuild commit: %v", buildCommit)

	app.Sugar.Infoln("before init config")

	configs, err := config.ReadConfig()
	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "read config")
	}
	err = config.Initialize(configs)
	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "read config")
	}

	return configs
}
