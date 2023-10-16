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
	RateLimit      int    `env:"RATE_LIMIT"`
	MemProfile     string `env:"MEM_PROFILE"`
	CryptoKey      string `env:"CRYPTO_KEY"`
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
	pflag.IntVarP(&agentFlags.RateLimit, "rateLimit", "l", 2,
		"Max count of parallel outbound requests to server")
	pflag.StringVarP(&agentFlags.MemProfile, "mem", "m", "base.pprof", "Path to file where mem stats will be saved")
	pflag.StringVarP(&agentFlags.CryptoKey, "crypto-key", "e", "/workspaces/metric-collector/cmd/keys/key.pub", "Path to public key RSA to encrypt messages")

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	fmt.Printf("\nRATE_LIMIT=%v", agentFlags.RateLimit)
	fmt.Printf("\nMEM_PROFILE=%v", agentFlags.MemProfile)
	fmt.Printf("\nCRYPTO_KEY=%v", agentFlags.CryptoKey)
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
	fmt.Printf("\nRATE_LIMIT=%v", agentFlags.RateLimit)
	fmt.Printf("\nMEM_PROFILE=%v", agentFlags.MemProfile)
	fmt.Printf("\nCRYPTO_KEY=%v", agentFlags.CryptoKey)

	return *agentFlags
}
