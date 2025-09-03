package config

import "os"

type ServiceConfig struct {
	UserServiceURL string
}

func LoadConfig() *ServiceConfig {
	return &ServiceConfig{
		UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
