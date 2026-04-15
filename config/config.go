package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	AdminUser            string
	AdminPass            string
	ReportsOutputDir     string
	ReportsRetentionDays int
	ReportsPublicURLBase string
	ReportsHTTPPort      string
}

func LoadConfig() Config {
	return Config{
		Port:                 getEnv("SERVER_PORT", "localhost:50051"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", "riccitelli-dev-secret-change-in-prod"),
		AdminUser:            getEnv("AUTH_ADMIN_USER", "admin"),
		AdminPass:            getEnv("AUTH_ADMIN_PASSWORD", "admin123"),
		ReportsOutputDir:     getEnv("REPORTS_OUTPUT_DIR", "./output/reports"),
		ReportsRetentionDays: getEnvInt("REPORTS_RETENTION_DAYS", 30),
		ReportsPublicURLBase: getEnv("REPORTS_PUBLIC_URL_BASE", "http://localhost:8080/reports/files"),
		ReportsHTTPPort:      getEnv("REPORTS_HTTP_PORT", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return defaultValue
}
