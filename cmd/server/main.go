package main

import (
	"fmt"
	app "metric-collector/cmd/server/app"
	store "metric-collector/internal/storage/memstorage"
	"net/http"

	"github.com/caarlos0/env/v8"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/pflag"
)

func main() {
	addr := initialize()
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

func initialize() app.ServerFlags {
	// try to get vars from env
	addr := new(app.ServerFlags)
	if err := env.Parse(addr); err != nil {
		panic(err)
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", addr.Address)
	// try to get vars from Flags
	if addr.Address == "" {
		pflag.StringVarP(&addr.Address, "addr", "a", "localhost:8080", "Net address host:port")
		pflag.Parse()
	}

	err := addr.Set(addr.Address)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", addr.Address)

	return *addr
}
