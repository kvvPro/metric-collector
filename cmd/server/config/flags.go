package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ServerFlags struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func Initialize() ServerFlags {
	// try to get vars from env
	srvFlags := new(ServerFlags)
	if err := env.Parse(srvFlags); err != nil {
		panic(err)
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", srvFlags.Address)
	fmt.Printf("STORE_INTERVAL=%v", srvFlags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", srvFlags.FileStoragePath)
	fmt.Printf("RESTORE=%v", srvFlags.Restore)
	// try to get vars from Flags
	if _, isSet := os.LookupEnv("ADDRESS"); !isSet {
		pflag.StringVarP(&srvFlags.Address, "addr", "a", "localhost:8080", "Net address host:port")
	}
	if _, isSet := os.LookupEnv("STORE_INTERVAL"); !isSet {
		pflag.IntVarP(&srvFlags.StoreInterval, "storeInterval", "i", 5,
			"Wait interval in seconds before dropping metrics to file")
	}
	if _, isSet := os.LookupEnv("FILE_STORAGE_PATH"); !isSet {
		pflag.StringVarP(&srvFlags.FileStoragePath, "fileStoragePath", "f", "/tmp/metrics-db.json",
			"Path to file where to save metrics")
	}
	if _, isSet := os.LookupEnv("RESTORE"); !isSet {
		pflag.BoolVarP(&srvFlags.Restore, "restore", "r", true,
			"True if restore values from file stored in FILE_STORAGE_PATH")
	}

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", srvFlags.Address)
	fmt.Printf("STORE_INTERVAL=%v", srvFlags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", srvFlags.FileStoragePath)
	fmt.Printf("RESTORE=%v", srvFlags.Restore)

	return *srvFlags
}
