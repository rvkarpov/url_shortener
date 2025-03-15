package config

import (
	"flag"
)

type Config struct {
	LaunchAddr  NetAddress
	PublishAddr string
}

func LoadConfig() *Config {
	var publishAddr VerifiedURL = "http://localhost:8080"
	flag.Var(&publishAddr, "b", "Result base address (format: valid URL)")

	launchAddr := NewNetAddress()
	flag.Var(&launchAddr, "a", "Launch address (format: host:port)")
	flag.Parse()

	return &Config{LaunchAddr: launchAddr, PublishAddr: string(publishAddr)}
}
