package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
	"syscall"
	"time"

	app "github.com/kvvPro/metric-collector/cmd/server/app"
	"github.com/kvvPro/metric-collector/cmd/server/config"

	"github.com/go-chi/chi/v5"
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
	app.Sugar = *logger.Sugar()

	app.Sugar.Infof("\nBuild version: %v", buildVersion)
	app.Sugar.Infof("\nBuild date: %v", buildDate)
	app.Sugar.Infof("\nBuild commit: %v", buildCommit)

	app.Sugar.Infoln("before init config")

	srvFlags, err := config.ReadConfig()
	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "read config")
	}
	err = config.Initialize(srvFlags)
	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "read config")
	}

	app.Sugar.Infoln("after init config")

	srv, err := app.NewServer(srvFlags)

	if err != nil {
		app.Sugar.Fatalw(err.Error(), "event", "create server")
	}

	app.Sugar.Infoln("before init config")

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	httpSrv := startServer(ctx, wg, srv, srvFlags)

	asyncCtx, cancelSaving := context.WithCancel(ctx)
	wg.Add(1)
	go srv.AsyncSaving(asyncCtx, wg)

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

	app.Sugar.Infoln("Try to save metrics...")
	err = srv.SaveToFile(ctx)
	if err != nil {
		app.Sugar.Infoln("Save to file failed: ", err.Error())
	}
	app.Sugar.Infoln("Metrics saved")

	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	app.Sugar.Infoln("Попытка мягко завершить сервер")
	if err := httpSrv.Shutdown(timeout); err != nil {
		app.Sugar.Errorf("Ошибка при попытке мягко завершить http-сервер: %v", err)
		// handle err
		if err = httpSrv.Close(); err != nil {
			app.Sugar.Errorf("Ошибка при попытке завершить http-сервер: %v", err)
		}
	}
	cancelSaving()
	wg.Wait()
	app.Sugar.Infoln("Server shutdown by signal: ", sigQuit)
}

func startServer(ctx context.Context, wg *sync.WaitGroup, srv *app.Server, srvFlags *config.ServerFlags) *http.Server {
	r := chi.NewMux()
	r.Use(srv.DecryptMiddleware,
		srv.CheckHashMiddleware,
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

	httpSrv := &http.Server{
		Addr:    srv.Address,
		Handler: r,
	}
	go func() {
		defer wg.Done()

		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			// записываем в лог ошибку, если сервер не запустился
			app.Sugar.Fatalw(err.Error(), "event", "start server")
		}
	}()

	return httpSrv
}
