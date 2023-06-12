package client

import (
	"errors"
	"strings"
)

type ClientFlags struct {
	Address        string `env:"ADDRESS"`
	Host           string
	Port           string
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
}

func (flags *ClientFlags) SetAddr(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	flags.Host = hp[0]
	flags.Port = hp[1]
	return nil
}
