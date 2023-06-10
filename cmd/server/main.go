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
	mux.Handle("/", http.HandlerFunc(srv.MainHandle))

	errs := http.ListenAndServe(":"+srv.Port, mux)
	if errs != nil {
		panic(errs)
	}
}
