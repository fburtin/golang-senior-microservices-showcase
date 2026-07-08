package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	MongoTimeout  time.Duration
}

func Load() Config {
	timeoutSeconds, err := strconv.Atoi(getEnv("MONGO_TIMEOUT_SECONDS", "10"))
	if err != nil {
		timeoutSeconds = 10
	}

	return Config{
		Port:          getEnv("PORT", "8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "go_showcase"),
		MongoTimeout:  time.Duration(timeoutSeconds) * time.Second,
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
