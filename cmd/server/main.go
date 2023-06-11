package main

import (
	app "metric-collector/cmd/server/app"
	store "metric-collector/internal/storage/memstorage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	storage := store.NewMemStorage()
	srv, err := app.NewServer(&storage, "8080")
	if err != nil {
		panic(err)
	}

	r := chi.NewMux()
	r.Handle("/update/*", http.HandlerFunc(srv.UpdateHandle))
	r.Handle("/value/*", http.HandlerFunc(srv.GetValueHandle))
	r.Handle("/", http.HandlerFunc(srv.AllMetricsHandle))

	errs := http.ListenAndServe(":"+srv.Port, r)
	if errs != nil {
		panic(errs)
	}
}
