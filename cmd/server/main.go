package main

import (
	"net/http"

	app "github.com/kvvPro/metric-collector/cmd/server/app"
	"github.com/kvvPro/metric-collector/cmd/server/config"
	store "github.com/kvvPro/metric-collector/internal/storage/memstorage"

	"github.com/go-chi/chi/v5"
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
	app.Sugar = *logger.Sugar()

	addr := config.Initialize()
	storage := store.NewMemStorage()
	srv := app.NewServer(&storage, addr.Host, addr.Port)

	r := chi.NewMux()
	r.Use(app.WithLogging)
	r.Handle("/update/", http.HandlerFunc(srv.UpdateJSONHandle))
	r.Handle("/update/*", http.HandlerFunc(srv.UpdateHandle))
	r.Handle("/value/*", http.HandlerFunc(srv.GetValueHandle))
	r.Handle("/value/", http.HandlerFunc(srv.GetValueJSONHandle))
	r.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))

	// записываем в лог, что сервер запускается
	app.Sugar.Infow(
		"Starting server",
		"addr", addr,
	)

	if err := http.ListenAndServe(srv.Host+":"+srv.Port, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		app.Sugar.Fatalw(err.Error(), "event", "start server")
	}
}
