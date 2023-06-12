package main

import (
	"metric-collector/cmd/agent/client"

	"github.com/spf13/pflag"
)

func main() {
	var faddr string
	agentFlags := new(client.ClientFlags)
	pflag.StringVarP(&faddr, "addr", "a", "localhost:8080", "Net address host:port")
	pflag.IntVarP(&agentFlags.ReportInterval, "reportInterval", "r", 10,
		"Wait interval in seconds before pushing metrics to server")
	pflag.IntVarP(&agentFlags.PollInterval, "pollInterval", "p", 2,
		"Wait interval in seconds before reading metrics")
	pflag.Parse()
	err := agentFlags.SetAddr(faddr)
	if err != nil {
		panic(err)
	}

	agent, err := client.NewClient(agentFlags.PollInterval, agentFlags.ReportInterval,
		agentFlags.Host, agentFlags.Port, "text/plain")
	if err != nil {
		panic(err)
	}

	agent.Run()
}
