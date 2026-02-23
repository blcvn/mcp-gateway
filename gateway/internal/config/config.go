package config

import (
	"os"
)

type Config struct {
	Port        string
	ToolsDir    string
	RedisURL    string
	Environment string
}

func Load() (*Config, error) {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		ToolsDir:    getEnv("TOOLS_DIR", "./tools"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		Environment: getEnv("ENV", "development"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
