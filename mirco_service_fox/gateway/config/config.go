package config

import "os"

type ServiceConfig struct {
	UserServiceURL         string
	ExperimentServiceURL   string
	NotificationServiceURL string
	SubmissionServiceURL   string
	GatewayPort            string
	JWTSecret              string
}

func LoadConfig() *ServiceConfig {
	return &ServiceConfig{
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		ExperimentServiceURL:   getEnv("EXPERIMENT_SERVICE_URL", "http://localhost:8082"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8083"),
		SubmissionServiceURL:   getEnv("SUBMISSION_SERVICE_URL", "http://localhost:8084"),
		GatewayPort:            getEnv("GATEWAY_PORT", "8080"),
		JWTSecret:              getEnv("JWT_SECRET", "a_secret_crect"),
	}
}

func getEnv(key, defaultValue string) string {
	// 实际项目中可以从环境变量或配置文件中读取
	// 这里简化为返回默认值
	value := os.Getenv(key)
	if value == "" {
		return defaultValue // 默认密钥，生产环境不建议使用默认值
	}

	return value
}
