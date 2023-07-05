package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ServerFlags struct {
	Address         string `env:"ADDRESS"`
	Host            string
	Port            string
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func (flags *ServerFlags) AddressToString() string {
	return flags.Host + ":" + flags.Port
}

func (flags *ServerFlags) SetAddress(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	flags.Host = hp[0]
	flags.Port = hp[1]
	return nil
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
	if srvFlags.Address == "" {
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
	err := srvFlags.SetAddress(srvFlags.Address)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", srvFlags.Address)
	fmt.Printf("STORE_INTERVAL=%v", srvFlags.StoreInterval)
	fmt.Printf("FILE_STORAGE_PATH=%v", srvFlags.FileStoragePath)
	fmt.Printf("RESTORE=%v", srvFlags.Restore)

	return *srvFlags
}