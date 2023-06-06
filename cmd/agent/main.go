package main

import (
	"metric-collector/cmd/agent/client"
)

func main() {
	agent, err := client.NewClient(2, 10, "http://localhost", "8080", "text/plain")
	if err != nil {
		panic(err)
	}

	agent.Run()
}
