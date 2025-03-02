package config

type Config struct {
	Host string
	Port uint16
}

func LoadConfig() Config {
	return Config{Host: "localhost", Port: 8080}
}
