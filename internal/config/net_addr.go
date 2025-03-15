package config

import (
	"errors"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func NewNetAddress() NetAddress {
	return NetAddress{Host: "localhost", Port: 8080}
}

func (addr NetAddress) String() string {
	return addr.Host + ":" + strconv.Itoa(addr.Port)
}

func (addr *NetAddress) Set(value string) error {
	hp := strings.Split(value, ":")
	if len(hp) != 2 {
		return errors.New("address in a form host:port required")
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}

	addr.Host = hp[0]
	addr.Port = port

	return nil
}
