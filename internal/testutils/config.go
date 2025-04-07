package testutils

import (
	"github.com/rvkarpov/url_shortener/internal/config"
)

func LoadTestConfig() config.Config {
	return config.Config{
		LaunchAddr:  config.NewNetAddress(),
		PublishAddr: "http://localhost:8080",
		ShortURLLen: 8,
	}
}
