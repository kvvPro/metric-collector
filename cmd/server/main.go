package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	rpprof "runtime/pprof"
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

	go startServer(context.Background(), srv, &srvFlags)
	go srv.AsyncSaving(context.Background())

	sigQuit := <-shutdown

	// создаём файл журнала профилирования памяти
	fmem, err := os.Create(srvFlags.MemProfile)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}

	app.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
	app.Sugar.Infoln("Try to save metrics...")
	err = srv.SaveToFile(context.Background())
	if err != nil {
		app.Sugar.Infoln("Save to file failed: ", err.Error())
	}
	app.Sugar.Infoln("Metrics saved")
}

func startServer(ctx context.Context, srv *app.Server, srvFlags *config.ServerFlags) {
	r := chi.NewMux()
	r.Use(srv.CheckHashMiddleware,
		app.GzipMiddleware,
		app.WithLogging)
	// r.Use(app.WithLogging)
	r.Handle("/ping", http.HandlerFunc(srv.PingHandle))
	r.Handle("/updates/", http.HandlerFunc(srv.UpdateBatchJSONHandle))
	r.Handle("/update/", http.HandlerFunc(srv.UpdateJSONHandle))
	r.Handle("/update/*", http.HandlerFunc(srv.UpdateHandle))
	r.Handle("/value/*", http.HandlerFunc(srv.GetValueHandle))
	r.Handle("/value/", http.HandlerFunc(srv.GetValueJSONHandle))
	r.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))
	r.Handle("/debug/pprof", http.HandlerFunc(pprof.Index))
	r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))

	// Manually add support for paths linked to by index page at /debug/pprof/
	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))

	app.Sugar.Infoln("before restoring values")

	srv.RestoreValues(ctx)

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
