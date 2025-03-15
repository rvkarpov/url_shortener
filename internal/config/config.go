package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	LaunchAddr  NetAddress  `env:"SERVER_ADDRESS,required"`
	PublishAddr VerifiedURL `env:"BASE_URL,required"`
}

func LoadConfig() *Config {
	log.Printf("Reading environment variables")
	cfg := &Config{LaunchAddr: NewNetAddress(), PublishAddr: "http://localhost:8080"}
	err := env.Parse(cfg)
	if err == nil {
		return cfg
	}

	log.Printf("Environment variables are not set properly: %s", err)
	log.Printf("Reading command line flags")

	flag.Var(&cfg.LaunchAddr, "a", "Launch address (format: host:port)")
	flag.Var(&cfg.PublishAddr, "b", "Result base address (format: valid URL)")
	flag.Parse()

	return cfg
}
