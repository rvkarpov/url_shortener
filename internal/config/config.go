package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	LaunchAddr   NetAddress  `env:"SERVER_ADDRESS"`
	PublishAddr  VerifiedURL `env:"BASE_URL"`
	StorageFile  string      `env:"FILE_STORAGE_PATH"`
	DBConnParams string      `env:"DATABASE_DSN"`
	TableName    string      `env:"DB_TABLE_NAME"`
	ShortURLLen  uint        `env:"SHORT_URL_LEN"`
}

func LoadConfig() *Config {
	log.Printf("Reading environment variables")
	cfg := &Config{
		LaunchAddr:  NewNetAddress(),
		PublishAddr: "http://localhost:8080",
	}

	flag.Var(&cfg.LaunchAddr, "a", "Launch address (format: host:port)")
	flag.Var(&cfg.PublishAddr, "b", "Result base address (format: valid URL)")
	flag.StringVar(&cfg.StorageFile, "f", "storage.dat", "Storage file path (format: filesystem path)")
	flag.StringVar(&cfg.DBConnParams, "d", "", "DB connection params (format: host=%s user=%s password=%s dbname=%s)")
	flag.StringVar(&cfg.TableName, "t", "urls", "DB table name (format: string)")
	flag.UintVar(&cfg.ShortURLLen, "l", 8, "short URL len (format: uint)")
	flag.Parse()

	env.Parse(cfg)

	return cfg
}
