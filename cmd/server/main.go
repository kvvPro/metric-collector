package main

import (
	app "metric-collector/cmd/server/app"
	store "metric-collector/internal/storage/memstorage"
	"net/http"
)

func main() {
	storage := store.NewMemStorage()
	srv, err := app.NewServer(&storage, "8080")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/update/", http.HandlerFunc(srv.UpdateHandle))
	mux.Handle("/value/", http.HandlerFunc(srv.GetValueHandle))
	mux.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))

	errs := http.ListenAndServe(":"+srv.Port, mux)
	if errs != nil {
		panic(errs)
	}
}
