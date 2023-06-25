package main

import (
	"github.com/kvvPro/metric-collector/cmd/agent/client"
	"github.com/kvvPro/metric-collector/cmd/agent/config"
)

func main() {
	agentFlags := config.Initialize()
	agent, err := client.NewClient(agentFlags.PollInterval, agentFlags.ReportInterval,
		agentFlags.Host, agentFlags.Port, "text/plain")
	if err != nil {
		panic(err)
	}

	agent.Run()
}
