package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ServerFlags struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DBConnection    string `env:"DATABASE_DSN"`
	HashKey         string `env:"KEY"`
}

func Initialize() ServerFlags {
	srvFlags := new(ServerFlags)
	// try to get vars from Flags
	// if _, isSet := os.LookupEnv("ADDRESS"); !isSet {
	pflag.StringVarP(&srvFlags.Address, "addr", "a", "localhost:8080", "Net address host:port")
	// }
	// if _, isSet := os.LookupEnv("DATABASE_DSN"); !isSet {
	// pflag.StringVarP(&srvFlags.DBConnection, "databaseConn", "d", "user=postgres password=postgres host=localhost port=5432 dbname=postgres sslmode=disable", "Connection string to DB: user=<> password=<> host=<> port=<> dbname=<>")
	pflag.StringVarP(&srvFlags.DBConnection, "databaseConn", "d", "", "Connection string to DB: user=<> password=<> host=<> port=<> dbname=<>")
	// }
	// if _, isSet := os.LookupEnv("STORE_INTERVAL"); !isSet {
	pflag.IntVarP(&srvFlags.StoreInterval, "storeInterval", "i", 5,
		"Wait interval in seconds before dropping metrics to file")
	// }
	// if _, isSet := os.LookupEnv("FILE_STORAGE_PATH"); !isSet {
	pflag.StringVarP(&srvFlags.FileStoragePath, "fileStoragePath", "f", "/tmp/metrics-db.json",
		"Path to file where to save metrics")
	// }
	// if _, isSet := os.LookupEnv("RESTORE"); !isSet {
	pflag.BoolVarP(&srvFlags.Restore, "restore", "r", true,
		"True if restore values from file stored in FILE_STORAGE_PATH")
	// }
	pflag.StringVarP(&srvFlags.HashKey, "hashKey", "k", "",
		"Hash key to calculate hash sum")

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", srvFlags.Address)
	fmt.Printf("STORE_INTERVAL=%v", srvFlags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", srvFlags.FileStoragePath)
	fmt.Printf("RESTORE=%v", srvFlags.Restore)
	fmt.Printf("DATABASE_DSN=%v", srvFlags.DBConnection)
	fmt.Printf("KEY=%v", srvFlags.HashKey)

	// try to get vars from env
	if err := env.Parse(srvFlags); err != nil {
		panic(err)
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", srvFlags.Address)
	fmt.Printf("STORE_INTERVAL=%v", srvFlags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", srvFlags.FileStoragePath)
	fmt.Printf("RESTORE=%v", srvFlags.Restore)
	fmt.Printf("DATABASE_DSN=%v", srvFlags.DBConnection)
	fmt.Printf("KEY=%v", srvFlags.HashKey)

	return *srvFlags
}
