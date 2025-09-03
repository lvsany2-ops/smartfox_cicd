package config

type ServiceConfig struct {
	UserServiceURL         string
	SubmissionServiceURL   string
	NotificationServiceURL string
}

func LoadConfig() *ServiceConfig {
	return &ServiceConfig{
		UserServiceURL:         getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		SubmissionServiceURL:   getEnv("SUBMISSION_SERVICE_URL", "http://localhost:8084"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8083"),
	}
}
