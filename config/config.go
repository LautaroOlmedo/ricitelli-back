package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	AdminUser   string
	AdminPass   string
}

func LoadConfig() Config {
	return Config{
		Port:        getEnv("SERVER_PORT", "localhost:50051"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "riccitelli-dev-secret-change-in-prod"),
		AdminUser:   getEnv("AUTH_ADMIN_USER", "admin"),
		AdminPass:   getEnv("AUTH_ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
