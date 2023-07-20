package config

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ClientFlags struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	HashKey        string `env:"KEY"`
}

func Initialize() ClientFlags {
	agentFlags := new(ClientFlags)
	// try to get vars from env
	if err := env.Parse(agentFlags); err != nil {
		panic(err)
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	// try to get vars from Flags
	if _, isSet := os.LookupEnv("ADDRESS"); !isSet {
		pflag.StringVarP(&agentFlags.Address, "addr", "a", "localhost:8080", "Net address host:port")
	}
	if _, isSet := os.LookupEnv("REPORT_INTERVAL"); !isSet {
		pflag.IntVarP(&agentFlags.ReportInterval, "reportInterval", "r", 10,
			"Wait interval in seconds before pushing metrics to server")
	}
	if _, isSet := os.LookupEnv("POLL_INTERVAL"); !isSet {
		pflag.IntVarP(&agentFlags.PollInterval, "pollInterval", "p", 2,
			"Wait interval in seconds before reading metrics")
	}
	if _, isSet := os.LookupEnv("KEY"); !isSet {
		pflag.StringVarP(&agentFlags.HashKey, "key", "k", "",
			"Hash key to calculate hash sum")
	}

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	fmt.Println()

	return *agentFlags
}
