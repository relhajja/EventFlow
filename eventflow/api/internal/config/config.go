package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Namespace   string
	JWTSecret   string
	LogLevel    string
	MetricsPort int
	DatabaseURL string
	NATSUrl     string
}

func Load() *Config {
	return &Config{
		Port:        getEnvAsInt("PORT", 8080),
		Namespace:   getEnv("NAMESPACE", "default"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		MetricsPort: getEnvAsInt("METRICS_PORT", 9090),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		NATSUrl:     getEnv("NATS_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
