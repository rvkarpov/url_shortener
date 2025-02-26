package config

type Config struct {
	Port uint16
}

func LoadConfig() Config {
	return Config{Port: 8080}
}
