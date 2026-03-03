package config

import "os"

type Config struct {
	Port string
}

func LoadConfig() Config {
	return Config{
		Port: getEnv("SERVER_PORT", "localhost:50051"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
