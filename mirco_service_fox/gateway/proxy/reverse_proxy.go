package proxy

import (
	"encoding/json"
	"fmt"
	"gateway/config"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetUrl, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// 设置超时
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	}
	// 修改请求头，确保正确传递
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		log.Printf("Proxying request to: %s", targetUrl.String())
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Header.Set("X-Original-URI", req.URL.String())

		// 传递用户信息（如果上游服务需要）
		if userID, exists := req.Context().Value(userIDKey).(string); exists {
			req.Header.Set("X-User-ID", userID)
		}
		if userRole, exists := req.Context().Value(userRoleKey).(string); exists {
			req.Header.Set("X-User-Role", userRole)
		}
		req.Host = targetUrl.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		errorMsg := fmt.Sprintf("Error occurred while proxying request: %v", err)
		log.Print(errorMsg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		response := map[string]string{
			"error":   "bad gateway",
			"message": "Unable to connect to upstream service",
			"details": errorMsg, // 添加错误详情到响应
		}
		jsonData, _ := json.Marshal(response)
		w.Write(jsonData)
	}
	return proxy, nil
}

// 根据路径确定目标服务
func GetTargetService(path string, cfg *config.ServiceConfig) string {
	// 这一处的代码转发需要根据实际路径进行修改
	switch {
	// 用户服务路由
	case strings.HasPrefix(path, "/api/auth") ||
		strings.HasPrefix(path, "/api/student_list") ||
		strings.HasPrefix(path, "/api/teacher/students") ||
		strings.HasPrefix(path, "/api/teacher/groups") ||
		strings.HasPrefix(path, "/internal/users"):
		return cfg.UserServiceURL

	// 通知服务路由
	case strings.HasPrefix(path, "/api/teacher/experiments/notifications") ||
		strings.HasPrefix(path, "/api/student/experiments/notifications"):
		return cfg.NotificationServiceURL

	// 提交服务路由
	case strings.Contains(path, "/save") ||
		strings.Contains(path, "/submit") ||
		strings.HasPrefix(path, "/api/student/submissions"):
		return cfg.SubmissionServiceURL

	// 实验服务路由
	case strings.HasPrefix(path, "/api/student/experiments") ||
		strings.HasPrefix(path, "/api/teacher/experiments") ||
		strings.HasPrefix(path, "/api/experiments"):
		return cfg.ExperimentServiceURL

	default:
		return cfg.UserServiceURL // 默认回退到用户服务
	}
}
