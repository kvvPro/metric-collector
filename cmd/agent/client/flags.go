package client

import (
	"errors"
	"strings"
)

type ClientFlags struct {
	Host           string
	Port           string
	ReportInterval int
	PollInterval   int
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
