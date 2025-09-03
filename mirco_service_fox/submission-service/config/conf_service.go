package config

import "os"

type ServiceConfig struct {
	ExperimentServiceURL string
	JudgeServiceURL      string
}

func LoadConfig() *ServiceConfig {
	return &ServiceConfig{
		ExperimentServiceURL: getEnv("EXPERIMENT_SERVICE_URL", "http://localhost:8082"),
		JudgeServiceURL:      getEnv("JUDGE_SERVICE_URL", "http://localhost:8085"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
