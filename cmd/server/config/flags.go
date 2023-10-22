package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ServerFlags struct {
	Address         string `env:"ADDRESS" json:"address"`
	StoreInterval   int    `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool   `env:"RESTORE" json:"restore"`
	DBConnection    string `env:"DATABASE_DSN" json:"database_dsn"`
	HashKey         string `env:"KEY" json:"hash_key"`
	MemProfile      string `env:"MEM_PROFILE" json:"mem_profile"`
	CryptoKey       string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config          string `env:"CONFIG" json:"config"`
}

func Initialize(flags *ServerFlags) error {
	// try to get vars from Flags
	pflag.StringVarP(&flags.Address, "addr", "a", "localhost:8080", "Net address host:port")
	// pflag.StringVarP(&flags.DBConnection, "databaseConn", "d", "user=postgres password=postgres host=localhost port=5432 dbname=postgres sslmode=disable", "Connection string to DB: user=<> password=<> host=<> port=<> dbname=<>")
	pflag.StringVarP(&flags.DBConnection, "databaseConn", "d", "", "Connection string to DB: user=<> password=<> host=<> port=<> dbname=<>")
	pflag.IntVarP(&flags.StoreInterval, "storeInterval", "i", 5,
		"Wait interval in seconds before dropping metrics to file")
	pflag.StringVarP(&flags.FileStoragePath, "fileStoragePath", "f", "/tmp/metrics-db.json",
		"Path to file where to save metrics")
	pflag.BoolVarP(&flags.Restore, "restore", "r", true,
		"True if restore values from file stored in FILE_STORAGE_PATH")
	pflag.StringVarP(&flags.HashKey, "hashKey", "k", "",
		"Hash key to calculate hash sum")
	pflag.StringVarP(&flags.MemProfile, "mem", "m", "base.pprof", "Path to file where mem stats will be saved")
	pflag.StringVarP(&flags.CryptoKey, "crypto-key", "e", "/workspaces/metric-collector/cmd/keys/key", "Path to private key RSA to decrypt messages")
	// pflag.StringVarP(&flags.Config, "config", "c", "/workspaces/metric-collector/cmd/server/config/config.json", "Path to server config file")

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", flags.Address)
	fmt.Printf("STORE_INTERVAL=%v", flags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", flags.FileStoragePath)
	fmt.Printf("RESTORE=%v", flags.Restore)
	fmt.Printf("DATABASE_DSN=%v", flags.DBConnection)
	fmt.Printf("KEY=%v", flags.HashKey)
	fmt.Printf("MEM_PROFILE=%v", flags.MemProfile)
	fmt.Printf("CRYPTO_KEY=%v", flags.CryptoKey)
	fmt.Printf("CONFIG=%v", flags.Config)

	// try to get vars from env
	if err := env.Parse(flags); err != nil {
		return err
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", flags.Address)
	fmt.Printf("STORE_INTERVAL=%v", flags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", flags.FileStoragePath)
	fmt.Printf("RESTORE=%v", flags.Restore)
	fmt.Printf("DATABASE_DSN=%v", flags.DBConnection)
	fmt.Printf("KEY=%v", flags.HashKey)
	fmt.Printf("MEM_PROFILE=%v", flags.MemProfile)
	fmt.Printf("CRYPTO_KEY=%v", flags.CryptoKey)
	fmt.Printf("CONFIG=%v", flags.Config)

	return nil
}

func ReadConfig() (*ServerFlags, error) {
	flags := new(ServerFlags)

	pflag.StringVarP(&flags.Config, "config", "c", "/workspaces/metric-collector/cmd/server/config/config.json", "Path to server config file")

	pflag.Parse()
	fmt.Printf("CONFIG=%v", flags.Config)

	data, err := os.ReadFile(flags.Config)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	if err := json.NewDecoder(reader).Decode(&flags); err != nil {
		return nil, err
	}

	return flags, nil
}
