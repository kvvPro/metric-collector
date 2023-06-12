package main

import (
	"fmt"
	"metric-collector/cmd/agent/client"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/spf13/pflag"
)

func main() {
	agentFlags := initialize()
	agent, err := client.NewClient(agentFlags.PollInterval, agentFlags.ReportInterval,
		agentFlags.Host, agentFlags.Port, "text/plain")
	if err != nil {
		panic(err)
	}

	agent.Run()
}

func initialize() client.ClientFlags {
	agentFlags := new(client.ClientFlags)
	// try to get vars from env
	if err := env.Parse(agentFlags); err != nil {
		panic(err)
	}
	fmt.Println("ENV-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nREPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)
	// try to get vars from Flags
	if agentFlags.Address == "" {
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

	pflag.Parse()
	err := agentFlags.SetAddr(agentFlags.Address)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nFLAGS-----------")
	fmt.Printf("ADDRESS=%v", agentFlags.Address)
	fmt.Printf("\nEPORT_INTERVAL=%v", agentFlags.ReportInterval)
	fmt.Printf("\nPOLL_INTERVAL=%v", agentFlags.PollInterval)

	return *agentFlags
}
