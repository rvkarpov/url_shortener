package config

import (
	"fmt"
	"net"
	"strconv"
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
	host, port_, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("address in a form host:port required")
	}

	port, err := strconv.Atoi(port_)
	if err != nil {
		return fmt.Errorf("an integer value is expected as the port value")
	}

	addr.Host = host
	addr.Port = port

	return nil
}

func (addr *NetAddress) UnmarshalText(text []byte) error {
	return addr.Set(string(text))
}
