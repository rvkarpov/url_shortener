package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	LaunchAddr   NetAddress  `env:"SERVER_ADDRESS"`
	PublishAddr  VerifiedURL `env:"BASE_URL"`
	StorageFile  string      `env:"FILE_STORAGE_PATH"`
	DBConnParams string      `env:"DATABASE_DSN"`
	TableName    string      `env:"DB_TABLE_NAME"`
	ShortURLLen  uint        `env:"SHORT_URL_LEN"`
	SecretKey    string      `env:"SECRET_KEY"`
}

func loadSecretKey() (string, error) {
	keyData, err := os.ReadFile("pseudo_secret_key_file.txt")
	if err != nil {
		return "pseudo_secret_key", nil
		//return "", fmt.Errorf("error reading secret key file: %w", err)
	}

	return string(keyData), nil
}

func LoadConfig() (*Config, error) {
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
	flag.StringVar(&cfg.SecretKey, "k", "", "DB table name (format: string)")
	flag.UintVar(&cfg.ShortURLLen, "l", 8, "short URL len (format: uint)")
	flag.Parse()

	env.Parse(cfg)

	if cfg.SecretKey == "" {
		secretKey, err := loadSecretKey()
		if err != nil {
			return nil, err
		}

		cfg.SecretKey = secretKey
	}

	return cfg, nil
}
