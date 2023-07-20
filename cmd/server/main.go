package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	app "github.com/kvvPro/metric-collector/cmd/server/app"
	"github.com/kvvPro/metric-collector/cmd/server/config"

	"github.com/go-chi/chi/v5"
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

	app.Sugar.Infoln("before init config")

	srvFlags := config.Initialize()

	app.Sugar.Infoln("after init config")

	srv, err := app.NewServer(srvFlags.Address,
		srvFlags.StoreInterval,
		srvFlags.FileStoragePath,
		srvFlags.Restore,
		srvFlags.DBConnection,
		srvFlags.HashKey)

	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "create server")
	}

	app.Sugar.Infoln("before init config")

	go startServer(srv, &srvFlags)
	go srv.AsyncSaving()

	sigQuit := <-shutdown
	app.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
	app.Sugar.Infoln("Try to save metrics...")
	err = srv.SaveToFile()
	if err != nil {
		app.Sugar.Infoln("Save to file failed: ", err.Error())
	}
	app.Sugar.Infoln("Metrics saved")
}

func startServer(srv *app.Server, srvFlags *config.ServerFlags) {
	r := chi.NewMux()
	r.Use(app.GzipMiddleware,
		app.WithLogging)
	// r.Use(app.WithLogging)
	r.Handle("/ping", http.HandlerFunc(srv.PingHandle))
	r.Handle("/updates/", http.HandlerFunc(srv.UpdateBatchJSONHandle))
	r.Handle("/update/", http.HandlerFunc(srv.UpdateJSONHandle))
	r.Handle("/update/*", http.HandlerFunc(srv.UpdateHandle))
	r.Handle("/value/*", http.HandlerFunc(srv.GetValueHandle))
	r.Handle("/value/", http.HandlerFunc(srv.GetValueJSONHandle))
	r.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))

	app.Sugar.Infoln("before restoring values")

	srv.RestoreValues()

	app.Sugar.Infoln("after restoring values")

	// записываем в лог, что сервер запускается
	app.Sugar.Infow(
		"Starting server",
		"srvFlags", srvFlags,
	)

	if err := http.ListenAndServe(srv.Address, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		app.Sugar.Fatalw(err.Error(), "event", "start server")
	}
}
