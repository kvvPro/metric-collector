package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ServerFlags struct {
	Address string `env:"ADDRESS"`
	Host    string
	Port    string
}

func (flags *ServerFlags) String() string {
	return flags.Host + ":" + flags.Port
}

func (flags *ServerFlags) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	flags.Host = hp[0]
	flags.Port = hp[1]
	return nil
}

func (flags *ServerFlags) Type() string {
	return "string"
}

func Initialize() ServerFlags {
	// try to get vars from env
	addr := new(ServerFlags)
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
