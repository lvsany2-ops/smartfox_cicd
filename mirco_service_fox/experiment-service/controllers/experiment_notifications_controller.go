package controllers

import (
	"bytes"
	"encoding/json"
	"experiment-service/config"
	"fmt"
	"net/http"
	"time"
)

// callNotificationService 调用通知服务的API
func callNotificationService(notificationData map[string]interface{}) error {
	// 通知服务的URL - 你需要根据实际部署情况修改这个URL
	cfg := config.LoadConfig()
	notificationServiceURL := fmt.Sprintf("%s/api/teacher/experiments/notifications", cfg.NotificationServiceURL)
	// 序列化请求数据
	jsonData, err := json.Marshal(notificationData)
	if err != nil {
		return fmt.Errorf("failed to marshal notification data: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", notificationServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification service returned status: %d", resp.StatusCode)
	}

	return nil
}
