package config

import (
	"fmt"

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

	// try to get vars from Flags
	pflag.StringVarP(&agentFlags.Address, "addr", "a", "localhost:8080", "Net address host:port")
	pflag.IntVarP(&agentFlags.ReportInterval, "reportInterval", "r", 10,
		"Wait interval in seconds before pushing metrics to server")
	pflag.IntVarP(&agentFlags.PollInterval, "pollInterval", "p", 2,
		"Wait interval in seconds before reading metrics")
	pflag.StringVarP(&agentFlags.HashKey, "key", "k", "",
		"Hash key to calculate hash sum")

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	fmt.Println()

	// try to get vars from env
	if err := env.Parse(agentFlags); err != nil {
		panic(err)
	}

	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)

	return *agentFlags
}
