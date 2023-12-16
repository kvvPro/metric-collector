package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

type ClientFlags struct {
	Address        string `env:"ADDRESS" json:"address"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	HashKey        string `env:"KEY" json:"hash_key"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
	MemProfile     string `env:"MEM_PROFILE" json:"mem_profile"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`
	ExchangeMode   string `env:"EXCHANGE_MODE" json:"exchange_mode"`
	Config         string `env:"CONFIG" json:"config"`
}

func Initialize(agentFlags *ClientFlags) error {

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
	pflag.StringVarP(&agentFlags.ExchangeMode, "exchange-mode", "x", "http", "Exchange mode - http or grpc")

	//pflag.StringVarP(&agentFlags.Config, "config", "c", "/workspaces/metric-collector/cmd/agent/config/config.json", "Path to agent config file")

	pflag.Parse()

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	fmt.Printf("\nRATE_LIMIT=%v", agentFlags.RateLimit)
	fmt.Printf("\nMEM_PROFILE=%v", agentFlags.MemProfile)
	fmt.Printf("\nCRYPTO_KEY=%v", agentFlags.CryptoKey)
	fmt.Printf("\nEXCHANGE_MODE=%v", agentFlags.ExchangeMode)
	fmt.Printf("\nCONFIG=%v", agentFlags.Config)
	fmt.Println()

	// try to get vars from env
	if err := env.Parse(agentFlags); err != nil {
		return err
	}

	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	fmt.Printf("\nKEY=%v", agentFlags.HashKey)
	fmt.Printf("\nRATE_LIMIT=%v", agentFlags.RateLimit)
	fmt.Printf("\nMEM_PROFILE=%v", agentFlags.MemProfile)
	fmt.Printf("\nCRYPTO_KEY=%v", agentFlags.CryptoKey)
	fmt.Printf("\nEXCHANGE_MODE=%v", agentFlags.ExchangeMode)
	fmt.Printf("\nCONFIG=%v", agentFlags.Config)

	return nil
}

func ReadConfig() (*ClientFlags, error) {
	flags := new(ClientFlags)

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
