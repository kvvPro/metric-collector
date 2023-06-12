package app

import (
	"errors"
	"strings"
)

type ServerFlags struct {
	Host string
	Port string
}

func (flags *ServerFlags) String() string {
	return flags.Host + ":" + flags.Port
}

func (flags *ServerFlags) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	flags.Host = hp[0]
	flags.Port = hp[1]
	return nil
}

func (flags *ServerFlags) Type() string {
	return "string"
}
