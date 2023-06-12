package main

import (
	app "metric-collector/cmd/server/app"
	store "metric-collector/internal/storage/memstorage"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/pflag"
)

func main() {
	// flags
	addr := new(app.ServerFlags)
	faddr := ""
	pflag.StringVarP(&faddr, "addr", "a", "localhost:8080", "Net address host:port")
	pflag.Parse()
	err := addr.Set(faddr)
	if err != nil {
		panic(err)
	}

	storage := store.NewMemStorage()
	srv, err := app.NewServer(&storage, addr.Host, addr.Port)
	if err != nil {
		panic(err)
	}

	r := chi.NewMux()
	r.Handle("/update/*", http.HandlerFunc(srv.UpdateHandle))
	r.Handle("/value/*", http.HandlerFunc(srv.GetValueHandle))
	r.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))

	errs := http.ListenAndServe(srv.Host+":"+srv.Port, r)
	if errs != nil {
		panic(errs)
	}
}
